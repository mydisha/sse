package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"time"
)

type Payment struct {
	PaymentID   int    `db:"payment_id" json:"-"`
	OrderMaskID string `db:"order_mask_id" json:"order_mask_id"`
	Status      bool   `db:"status" json:"status"`
}

type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type SSEHandler struct {
	clients        map[chan string]interface{}
	newClients     chan chan string
	defunctClients chan chan string
	messages       chan string
}

func NewSSEHandler() *SSEHandler {
	b := &SSEHandler{
		clients:        make(map[chan string]interface{}),
		newClients:     make(chan (chan string)),
		defunctClients: make(chan (chan string)),
		messages:       make(chan string, 10),
	}
	return b
}

func (b *SSEHandler) HandleEvents() {
	go func() {
		for {
			select {
			case s := <-b.newClients:
				b.clients[s] = true
			case s := <-b.defunctClients:
				delete(b.clients, s)
				close(s)
			case msg := <-b.messages:
				for s, _ := range b.clients {
					s <- msg
				}
			}
		}
	}()
}

func (b *SSEHandler) SendString(msg string) {
	b.messages <- msg
}

func (b *SSEHandler) SendJSON(obj interface{}) {
	tmp, err := json.Marshal(obj)
	if err != nil {
		log.Panic("Error while sending JSON object:", err)
	}
	b.messages <- string(tmp)
}

func (b *SSEHandler) Subscribe(c *gin.Context) {
	db, err := sqlx.Connect("mysql", "user:password@(localhost:3306)/sse")
	if err != nil {
		panic(err.Error())
	}

	w := c.Writer
	f, ok := w.(http.Flusher)
	if !ok {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Streaming unsupported"))
		return
	}

	messageChan := make(chan string)

	b.newClients <- messageChan

	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		b.defunctClients <- messageChan
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	go func() {
		for {
			var payment Payment
			err = db.Get(&payment, "SELECT * FROM payments WHERE order_mask_id = ? AND status = ?", "QEO2LD", true)
			if err != nil {
				fmt.Println(err.Error())
			}

			if payment.Status {
				b.SendJSON(Message{
					Event: "poll",
					Data:  payment,
				})
				break
			}

			time.Sleep(time.Second * 2)
		}
	}()

	for {
		msg, open := <-messageChan
		if !open {
			break
		}

		fmt.Printf("message %s", msg)
		fmt.Fprintf(w, "data: Message: %s\n\n", msg)

		f.Flush()
	}

	c.AbortWithStatus(http.StatusOK)
}

func main() {
	sse := NewSSEHandler()
	sse.HandleEvents()

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/events", sse.Subscribe)

	r.Run(":8000")
}
