package database

import (
	"context"
	"database/sql" // добавлено
	"fmt"
	"log"

	"github.com/Revachol/WB_L0_microservice/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func DbOrders() {
	db := New("postgres", "secret", "mydb", "localhost", 5432)

	// получим заказы
	orders, err := db.GetOrders()
	if err != nil {
		log.Fatalf("Ошибка выборки: %v", err)
	}

	for _, o := range orders {
		fmt.Printf("ID=%s DATA=%s\n", o.ID, o.Data)
	}
}

type DB struct {
	conn *sqlx.DB
}

type Order struct {
	ID   string `db:"id"`
	Data string `db:"data"`
}

func New(user, password, dbname, host string, port int) *DB {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		"admin", "root", "orders_db", "localhost", 5433)

	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	return &DB{conn: conn}
}

func (db *DB) SQL() *sql.DB {
	return db.conn.DB
}

func (db *DB) GetOrders() ([]Order, error) {
	var orders []Order
	err := db.conn.Select(&orders, "SELECT id, data FROM orders")
	return orders, err
}

func (db *DB) InsertOrder(order Order) error {
	_, err := db.conn.NamedExec(
		`INSERT INTO orders (id, data) VALUES (:id, :data)`,
		order,
	)
	return err
}

func (db *DB) GetOrderByID(id string) (Order, error) {
	var order Order
	err := db.conn.Get(&order, "SELECT id, data FROM orders WHERE id=$1", id)
	return order, err
}

func (db *DB) GetFullOrderByUID(oid string) (models.Order, error) {
	var order models.Order
	err := db.conn.Get(&order, `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из orders: %v", err)
		return order, err
	}

	err = db.conn.Get(&order.Delivery, `
		SELECT name, phone, zip, city, address, region, email 
		FROM deliveries WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из deliveries: %v", err)
		return order, err
	}

	err = db.conn.Get(&order.Payment, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee 
		FROM payments WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из payments: %v", err)
		return order, err
	}

	err = db.conn.Select(&order.Items, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status 
		FROM items WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из items: %v", err)
		return order, err
	}

	return order, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func GetFullOrderByID(ctx context.Context, db *sql.DB, orderUID string) (models.Order, error) {
	var order models.Order

	err := db.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return order, err
	}

	err = db.QueryRowContext(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM deliveries WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return order, err
	}

	err = db.QueryRowContext(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
		       bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		return order, err
	}

	rows, err := db.QueryContext(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size,
		       total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return order, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return order, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
