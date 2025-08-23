package Internal

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
)

const (
	ordersTable   = "orders"
	deliveryTable = "delivery"
	paymentTable  = "payments"
	itemTable     = "items"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) RunMigrations() error {
	migrationSQL, err := os.ReadFile("schema/001_init_schema.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = s.db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func (s *Storage) Create(order Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createOrderQuery := fmt.Sprintf(`
        INSERT INTO %s (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		ordersTable)

	_, err = tx.Exec(createOrderQuery,
		order.Orders.OrderUid,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("into orders failed: %w", err)
	}

	createPaymentQuery := fmt.Sprintf(`
        INSERT INTO %s (
            transaction, request_id, currency, provider, amount, 
            payment_dt, bank, delivery_cost, goods_total, custom_fee, order_uid
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		paymentTable)

	_, err = tx.Exec(createPaymentQuery,
		order.Transaction,
		order.RequestID,
		order.Currency,
		order.Provider,
		order.Amount,
		order.PaymentDt,
		order.Bank,
		order.DeliveryCost,
		order.GoodsTotal,
		order.CustomFee,
		order.Orders.OrderUid,
	)
	if err != nil {
		return fmt.Errorf("into payments failed: %w", err)
	}

	createDeliveryQuery := fmt.Sprintf(`
        INSERT INTO %s (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		deliveryTable)

	_, err = tx.Exec(createDeliveryQuery,
		order.Orders.OrderUid,
		order.Name,
		order.Phone,
		order.Zip,
		order.City,
		order.Address,
		order.Region,
		order.Email,
	)
	if err != nil {
		return fmt.Errorf("into delivery failed: %w", err)
	}

	createItemQuery := fmt.Sprintf(`
        INSERT INTO %s (
            order_uid, chrt_id, track_number, price, rid, name, 
            sale, size, total_price, nm_id, brand, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		itemTable)

	for _, item := range order.Items {
		_, err = tx.Exec(createItemQuery,
			order.Orders.OrderUid,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("into items failed: %w", err)
		}
	}
	log.Println("Order created successfully: %s", order.Orders.OrderUid)
	return tx.Commit()
}

func (s *Storage) GetFromDb(OrderUid string) (Order, error) {
	var order Order

	query := `SELECT * FROM orders WHERE order_uid = $1`
	err := s.db.Get(&order.Orders, query, OrderUid)
	if err != nil {
		return order, fmt.Errorf("failed to get order: %w", err)
	}

	paymentQuery := `SELECT * FROM payments WHERE order_uid = $1`
	err = s.db.Get(&order.Payment, paymentQuery, OrderUid)
	if err != nil {
		return order, fmt.Errorf("failed to get payment: %w", err)
	}

	deliveryQuery := `SELECT * FROM delivery WHERE order_uid = $1`
	err = s.db.Get(&order.Delivery, deliveryQuery, OrderUid)
	if err != nil {
		return order, fmt.Errorf("failed to get delivery: %w", err)
	}

	itemsQuery := `SELECT * FROM items WHERE order_uid = $1`
	err = s.db.Select(&order.Items, itemsQuery, OrderUid)
	if err != nil {
		return order, fmt.Errorf("failed to get items: %w", err)
	}

	log.Println("Order отправлен: %s", order.Orders.OrderUid)
	return order, nil
}
