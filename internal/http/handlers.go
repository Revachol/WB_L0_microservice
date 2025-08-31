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
	// Новый endpoint для debug: получение заказов и вывод их в терминал
	mux.HandleFunc("/debug/orders", DebugOrdersHandler)

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
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	log.Println("Получен order_uid:", orderUID)

	db := database.New("", "", "", "", 0)
	fullOrder, err := db.GetFullOrderByUID(orderUID)
	if err != nil {
		log.Printf("Ошибка получения заказа по order_uid %s: %v", orderUID, err)
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}
	// Рендерим HTML-шаблон вместо вывода JSON
	renderTemplate(w, "order.html", fullOrder)
}

// Новый обработчик для получения всех заказов и вывода их в лог
func DebugOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// Создаём подключение к БД (параметры не используются, т.к. внутри New захардкожены)
	db := database.New("", "", "", "", 0)
	orders, err := db.GetOrders()
	if err != nil {
		log.Printf("Ошибка получения заказов: %v", err)
		http.Error(w, "Ошибка получения заказов", http.StatusInternalServerError)
		return
	}
	for _, o := range orders {
		log.Printf("Заказ ID: %s, Data: %s", o.ID, o.Data)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Заказы залогированы в терминале"))
}
