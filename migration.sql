CREATE DATABASE sse;

CREATE TABLE sse.payments (
                              payment_id BIGINT auto_increment NOT NULL,
                              order_mask_id varchar(30) NOT NULL,
                              status BOOL DEFAULT false NOT NULL,
                              CONSTRAINT payments_PK PRIMARY KEY (payment_id)
)
    ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;