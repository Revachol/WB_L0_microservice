KAFKA_BROKER=localhost:9092
TOPIC=orders
run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

send:
	@if [ -z "$(file)" ]; then \
		echo "❌ Укажите файл: make send file=data/1correct_model.json"; \
		exit 1; \
	fi; \
	go run cmd/producer/main.go $${file}
