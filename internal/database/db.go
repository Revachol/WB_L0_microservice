package database

import (
	"database/sql" // добавлено
	"fmt"
	"log"

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

// Close закрывает соединение с БД
func (db *DB) Close() error {
	return db.conn.Close()
}
