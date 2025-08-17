package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, err := template.ParseFiles(filepath.Join("web/templates", tmpl))
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка рендера", http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func orderRedirectHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	http.Redirect(w, r, "/order/"+orderID, http.StatusSeeOther)
}

// обработчик для /order/<id>
func orderHandler(w http.ResponseWriter, r *http.Request) {
	// вырезаем "/order/"
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	log.Println("Получен order_id:", orderID)

	data := map[string]string{
		"OrderID": orderID,
	}
	renderTemplate(w, "order.html", data)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/order", orderRedirectHandler)
	mux.HandleFunc("/order/", orderHandler)
	log.Println("Сервер запущен на http://localhost:" + port)
	http.ListenAndServe(":"+port, mux)
}
