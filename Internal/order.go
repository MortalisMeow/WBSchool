package Internal

import (
	"time"
)

type Order struct {
	Orders   `json:"orders"`
	Payment  `json:"payment"`
	Delivery `json:"deliveries"`
	Items    []Item `json:"items"`
}

type Orders struct {
	OrderUid          string    `json:"order_uid" db:"order_uid"`
	TrackNumber       string    `json:"track_number" db:"track_number"`
	Entry             string    `json:"entry" db:"entry"`
	Locale            string    `json:"locale" db:"locale"`
	InternalSignature *string   `json:"internal_signature,omitempty" db:"internal_signature"`
	CustomerID        string    `json:"customer_id" db:"customer_id"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service"`
	Shardkey          string    `json:"shardkey" db:"shardkey"`
	SmID              int       `json:"sm_id" db:"sm_id"`
	DateCreated       time.Time `json:"date_created" db:"date_created"`
	OofShard          string    `json:"oof_shard" db:"oof_shard"`
}

type Payment struct {
	PaymentID    int    `json:"-" db:"payment_id"`
	Transaction  string `json:"transaction" db:"transaction"`
	RequestID    string `json:"request_id,omitempty" db:"request_id"`
	Currency     string `json:"currency" db:"currency"`
	Provider     string `json:"provider" db:"provider"`
	Amount       int    `json:"amount" db:"amount"`
	PaymentDt    int64  `json:"payment_dt" db:"payment_dt"`
	Bank         string `json:"bank" db:"bank"`
	DeliveryCost int64  `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee"`
	OrderUID     string `json:"-" db:"order_uid"`
}

type Delivery struct {
	DeliveryID int    `json:"-" db:"delivery_id"`
	OrderUID   string `json:"-" db:"order_uid"`
	Name       string `json:"name" db:"name"`
	Phone      string `json:"phone" db:"phone"`
	Zip        string `json:"zip" db:"zip"`
	City       string `json:"city" db:"city"`
	Address    string `json:"address" db:"address"`
	Region     string `json:"region" db:"region"`
	Email      string `json:"email" db:"email"`
}

type Item struct {
	ChrtID      int64  `json:"chrt_id" db:"chrt_id"`
	TrackNumber string `json:"track_number" db:"track_number"`
	Price       int64  `json:"price" db:"price"`
	Rid         string `json:"rid" db:"rid"`
	Name        string `json:"name" db:"name"`
	Sale        int64  `json:"sale" db:"sale"`
	Size        string `json:"size" db:"size"`
	TotalPrice  int64  `json:"total_price" db:"total_price"`
	NmID        int64  `json:"nm_id" db:"nm_id"`
	Brand       string `json:"brand" db:"brand"`
	Status      int    `json:"status" db:"status"`
	OrderUID    string `json:"-" db:"order_uid"`
}
