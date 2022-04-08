package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type Payment struct {
	PaymentID   int    `db:"payment_id" json:"-"`
	OrderMaskID string `db:"order_mask_id" json:"order_mask_id"`
	Status      bool   `db:"status" json:"status"`
}

func main() {
	fmt.Println("Short poll example")

	db, err := sqlx.Connect("mysql", "username:password@(localhost:3306)/sse")
	if err != nil {
		panic(err.Error())
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/insert", func(c *gin.Context) {
		var payment Payment
		err = c.ShouldBindJSON(&payment)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.MustExec("INSERT INTO payments (order_mask_id, status) VALUES (?, ?)", payment.OrderMaskID, payment.Status)
		c.JSON(http.StatusOK, gin.H{
			"message": "Inserted",
		})
	})

	r.GET("/pool", func(c *gin.Context) {
		orderMaskId := c.Query("orderMaskId")

		var payment Payment
		err = db.Get(&payment, "SELECT * FROM payments WHERE order_mask_id = ?", orderMaskId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, payment)
	})

	r.GET("/callback", func(c *gin.Context) {
		orderMaskId := c.Query("orderMaskId")

		_, err = db.Exec("UPDATE payments SET status = 1 WHERE order_mask_id = ?", orderMaskId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	defer db.Close()

	err = r.Run(":8080")
	if err != nil {
		panic(err.Error())
	}
}
