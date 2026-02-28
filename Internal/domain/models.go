package domain

import (
	"time"
)

type Order struct {
	Orders   `json:"orders"`
	Payment  `json:"payment"`
	Delivery `json:"delivery"`
	Items    []Item `json:"items" fakesize:"1,3"`
}

type Orders struct {
	OrderUid          string    `json:"order_uid" db:"order_uid" fake:"{uuid}"`
	TrackNumber       string    `json:"track_number" db:"track_number" fake:"WB{digit:10}TEST"`
	Entry             string    `json:"entry" db:"entry" fake:"WBIL"`
	Locale            string    `json:"locale" db:"locale" fake:"en"`
	InternalSignature *string   `json:"internal_signature,omitempty" db:"internal_signature" fake:"{password:10}"`
	CustomerID        string    `json:"customer_id" db:"customer_id" fake:"cust-{number:100,999}"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" fake:"meest"`
	Shardkey          string    `json:"shardkey" db:"shardkey" fake:"{number:1,10}"`
	SmID              int       `json:"sm_id" db:"sm_id" fake:"{number:1,100}"`
	DateCreated       time.Time `json:"date_created" db:"date_created" fake:"{date}"`
	OofShard          string    `json:"oof_shard" db:"oof_shard" fake:"{number:1,10}"`
}

type Payment struct {
	PaymentID    int    `json:"-" db:"payment_id" fake:"-"`
	Transaction  string `json:"transaction" db:"transaction" fake:"{uuid}"`
	RequestID    string `json:"request_id,omitempty" db:"request_id" fake:"req-{number:100,999}"`
	Currency     string `json:"currency" db:"currency" fake:"USD"`
	Provider     string `json:"provider" db:"provider" fake:"wbpay"`
	Amount       int    `json:"amount" db:"amount" fake:"{number:1000,5000}"`
	PaymentDt    int64  `json:"payment_dt" db:"payment_dt" fake:"-"`
	Bank         string `json:"bank" db:"bank" fake:"alpha"`
	DeliveryCost int64  `json:"delivery_cost" db:"delivery_cost" fake:"{number:500,1500}"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total" fake:"{number:500,2000}"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee" fake:"0"`
	OrderUid     string `json:"-" db:"order_uid"`
}

type Delivery struct {
	DeliveryID int    `json:"-" db:"delivery_id" fake:"-"`
	OrderUid   string `json:"-" db:"order_uid"`
	Name       string `json:"name" db:"name" fake:"{person.name}"`
	Phone      string `json:"phone" db:"phone" fake:"{phone}"`
	Zip        string `json:"zip" db:"zip" fake:"{zip}"`
	City       string `json:"city" db:"city" fake:"{city}"`
	Address    string `json:"address" db:"address" fake:"{street} {streetnumber}"`
	Region     string `json:"region" db:"region" fake:"{state}"`
	Email      string `json:"email" db:"email" fake:"{email}"`
}

type Item struct {
	ID          int64  `json:"-" db:"id" fake:"-"`
	ChrtID      int64  `json:"chrt_id" db:"chrt_id" fake:"{number:1000,9999}"`
	TrackNumber string `json:"track_number" db:"track_number"`
	Price       int64  `json:"price" db:"price" fake:"{number:100,2000}"`
	Rid         string `json:"rid" db:"rid" fake:"rid-{uuid}"`
	Name        string `json:"name" db:"name" fake:"{product.name}"`
	Sale        int64  `json:"sale" db:"sale" fake:"{number:0,30}"`
	Size        string `json:"size" db:"size" fake:"{number:30,50}"`
	TotalPrice  int64  `json:"total_price" db:"total_price" fake:"{number:100,5000}"`
	NmID        int64  `json:"nm_id" db:"nm_id" fake:"{number:1000,9999}"`
	Brand       string `json:"brand" db:"brand" fake:"{company}"`
	Status      int    `json:"status" db:"status" fake:"202"`
	OrderUid    string `json:"-" db:"order_uid"`
}
