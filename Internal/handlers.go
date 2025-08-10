package Internal

import (
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
)

func HomePage(w http.ResponseWriter, _ *http.Request) {
	htmlTemplate, err := template.ParseFiles("./ui/html/index.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
	err = htmlTemplate.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
	//tmpl := template.Must(template.ParseFiles("ui/html/info.html"))
	//tmpl.Execute(w, nil)

}

func GetOrderId(w http.ResponseWriter, r *http.Request) string {
	OrderId := r.URL.Query().Get("order_uid")

	return OrderId
	//получаем ордер айди
	//проверка на неправильные значения
	//проверяем в кэше
	//если нет в кэше идем в бд
	//логика отдачи в отрисовку данных из json
}

func (s *Storage) GetFromDb(OrderId string) {

	order := &Order{}

	query := `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale,
            o.internal_signature, o.customer_id, o.delivery_service,
            o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            
            p.transaction, p.request_id, p.currency, p.provider,
            p.amount, p.payment_dt, p.bank, p.delivery_cost,
            p.goods_total, p.custom_fee,
            
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email
        FROM orders o
        LEFT JOIN payments p ON o.order_uid = p.order_uid
        LEFT JOIN deliveries d ON o.order_uid = d.order_uid
        WHERE o.order_uid = $1`

	err := s.Db.QueryRow(query, OrderId).Scan(
		&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)

	items, err := s.getOrderItems(OrderId)
	order.Items = items

	if err != nil {
		return
	}
	return
}

func (s *Storage) getOrderItems(OrderId string) ([]Item, error) {
	query := `
        SELECT 
            chrt_id, price, rid, name, sale, size,
            total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1`

	rows, err := s.Db.Query(query, OrderId)
	if err != nil {
		return nil, fmt.Errorf("Ошибка обработки items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ChrtID, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return nil, fmt.Errorf("Не удалось сканировать items: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func GetOrder(w http.ResponseWriter, r *http.Request) {}
