package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Revachol/WB_L0_microservice/internal/database"
	apphttp "github.com/Revachol/WB_L0_microservice/internal/http"
	"github.com/Revachol/WB_L0_microservice/internal/kafka"
)

func main() {
	// Получаем объект подключения через database.New
	dbObj := database.New("", "", "", "", 0)
	defer dbObj.Close()

	// Запускаем HTTP-сервер в горутине
	go apphttp.HttpServer()

	// Создаём контекст и запускаем consumer с подключением к БД
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewConsumer([]string{"localhost:9092"}, "orders", "my-group", dbObj.SQL())
	go consumer.Run(ctx)

	// Ожидаем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Выключаем сервис...")
}
