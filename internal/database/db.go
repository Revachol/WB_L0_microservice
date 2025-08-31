package database

import (
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

// Order структура соответствует таблице orders
type Order struct {
	ID   string `db:"id"`
	Data string `db:"data"` // можно хранить JSON как строку, а потом распарсить
}

// New создаёт новое подключение к БД
func New(user, password, dbname, host string, port int) *DB {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		"admin", "root", "orders_db", "localhost", 5433)

	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	return &DB{conn: conn}
}

// SQL возвращает «сырое» подключение *sql.DB
func (db *DB) SQL() *sql.DB {
	return db.conn.DB // исправлено: убраны скобки
}

// GetOrders возвращает все заказы
func (db *DB) GetOrders() ([]Order, error) {
	var orders []Order
	err := db.conn.Select(&orders, "SELECT id, data FROM orders")
	return orders, err
}

// InsertOrder вставляет новый заказ
func (db *DB) InsertOrder(order Order) error {
	_, err := db.conn.NamedExec(
		`INSERT INTO orders (id, data) VALUES (:id, :data)`,
		order,
	)
	return err
}

// GetOrderByID возвращает заказ по его id
func (db *DB) GetOrderByID(id string) (Order, error) {
	var order Order
	err := db.conn.Get(&order, "SELECT id, data FROM orders WHERE id=$1", id)
	return order, err
}

// GetFullOrderByUID возвращает заказ со всеми связанными данными по order_uid
func (db *DB) GetFullOrderByUID(oid string) (models.Order, error) {
	var order models.Order
	// Получаем данные из таблицы orders
	err := db.conn.Get(&order, `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из orders: %v", err)
		return order, err
	}

	// Получаем данные из таблицы deliveries
	err = db.conn.Get(&order.Delivery, `
		SELECT name, phone, zip, city, address, region, email 
		FROM deliveries WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из deliveries: %v", err)
		return order, err
	}

	// Получаем данные из таблицы payments
	err = db.conn.Get(&order.Payment, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee 
		FROM payments WHERE order_uid=$1
	`, oid)
	if err != nil {
		log.Printf("Ошибка выборки из payments: %v", err)
		return order, err
	}

	// Получаем данные из таблицы items
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

// Close закрывает соединение с БД
func (db *DB) Close() error {
	return db.conn.Close()
}
