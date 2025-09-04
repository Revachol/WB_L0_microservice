package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("❌ Укажите json-файл: go run cmd/producer/main.go data/order1.json")
	}
	file := os.Args[1]

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("не удалось прочитать файл %s: %v", file, err)
	}

	log.Printf("Отправка данных из файла %s в Kafka...", file)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders",
	})
	defer writer.Close()

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: data,
	})
	if err != nil {
		log.Fatalf("ошибка при отправке сообщения: %v", err)
	}

	fmt.Printf("✅ Сообщение из %s отправлено в Kafka\n", file)
}
