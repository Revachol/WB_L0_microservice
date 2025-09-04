package http

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Revachol/WB_L0_microservice/internal/cache"
	"github.com/Revachol/WB_L0_microservice/internal/database"
)

func HttpServer(cache *cache.Cache, db *sql.DB) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// обработчики
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/order", OrderRedirectHandler)
	mux.HandleFunc("/order/", OrderHandler(cache))
	mux.HandleFunc("/order_db/", OrderHandlerDB(db))

	log.Println("Сервер запущен на http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, err := template.ParseFiles(filepath.Join("web/templates", tmpl))
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Ошибка рендера шаблона: %v", err)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func OrderRedirectHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	http.Redirect(w, r, "/order/"+orderID, http.StatusSeeOther)
}

func OrderHandler(cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := strings.TrimPrefix(r.URL.Path, "/order/")
		order, ok := cache.Get(id)
		if !ok {
			log.Printf("Order %s not found in cache", id)
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		duration := time.Since(start)
		log.Printf("Order %s served from CACHE in %v ✅", id, duration)

		renderTemplate(w, "order.html", order)
	}
}

func OrderHandlerDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := strings.TrimPrefix(r.URL.Path, "/order_db/")
		order, err := database.GetFullOrderByID(r.Context(), db, id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		duration := time.Since(start)
		log.Printf("Order %s served from DB in %v 🐢", id, duration)

		renderTemplate(w, "order.html", order)
	}
}
