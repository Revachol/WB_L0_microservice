package cache

import (
	"database/sql"
	"log"
	"sync"

	"github.com/Revachol/WB_L0_microservice/internal/models"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]models.Order
}

func NewCache(db *sql.DB) *Cache {
	c := &Cache{
		orders: make(map[string]models.Order),
	}

	rows, err := db.Query(`
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		JOIN deliveries d ON o.order_uid = d.order_uid
		JOIN payments p ON o.order_uid = p.order_uid
	`)
	if err != nil {
		log.Printf("failed to load orders from DB: %v", err)
		return c
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			log.Printf("failed to scan order: %v", err)
			continue
		}

		itemRows, err := db.Query(`
			SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
			FROM items WHERE order_uid = $1
		`, order.OrderUID)
		if err != nil {
			log.Printf("failed to load items for order %s: %v", order.OrderUID, err)
			continue
		}

		for itemRows.Next() {
			var item models.Item
			if err := itemRows.Scan(
				&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
				&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
			); err != nil {
				log.Printf("failed to scan item: %v", err)
				continue
			}
			order.Items = append(order.Items, item)
		}
		itemRows.Close()

		c.orders[order.OrderUID] = order
	}

	log.Printf("cache initialized with %d orders", len(c.orders))
	return c
}

// Add добавляет заказ в кэш
func (c *Cache) Add(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[orderUID]
	return order, ok
}

func (c *Cache) GetAll() []models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]models.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, order)
	}
	return orders
}
