package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Revachol/WB_L0_microservice/internal/cache"
	"github.com/Revachol/WB_L0_microservice/internal/database"
	apphttp "github.com/Revachol/WB_L0_microservice/internal/http"
	"github.com/Revachol/WB_L0_microservice/internal/kafka"
)

func main() {
	dbObj := database.New("", "", "", "", 0)
	cache := cache.NewCache(dbObj.SQL())
	defer dbObj.Close()

	go apphttp.HttpServer(cache, dbObj.SQL())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafka.NewConsumer([]string{"localhost:9092"}, "orders", "my-group", dbObj.SQL(), cache)
	go consumer.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Выключаем сервис...")
}
