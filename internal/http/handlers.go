package http

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Revachol/WB_L0_microservice/internal/database" // добавлен импорт БД
)

func HttpServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	// статика
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// обработчики
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/order", OrderRedirectHandler)
	mux.HandleFunc("/order/", OrderHandler)

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
		// логируем ошибку вместо повторного вызова http.Error
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

func OrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	log.Println("Получен order_id:", orderID)

	// Создаём подключение к БД (параметры не используются, т.к. внутри New захардкожены)
	db := database.New("", "", "", "", 0)
	order, err := db.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Выводим данные заказа как JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(order.Data))
}
