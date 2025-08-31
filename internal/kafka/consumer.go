package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/Revachol/WB_L0_microservice/internal/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	db     *sql.DB
}

func NewConsumer(brokers []string, topic, groupID string, db *sql.DB) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader: r,
		db:     db,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid json: %v", err)
			continue
		}

		if err := c.saveOrder(ctx, order); err != nil {
			log.Printf("failed to save order: %v", err)
			continue
		}

		log.Printf("order %s saved", order.OrderUID)
	}
}

func (c *Consumer) saveOrder(ctx context.Context, order models.Order) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. orders
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders 
		(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	// 2. deliveries
	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries 
		(order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	// 3. payments
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments 
		(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee)
	if err != nil {
		return err
	}

	// 4. items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items 
			(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
