package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"url-shortener/config"
	postgres "url-shortener/db"
)

var store *postgres.URLStore

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request config.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	shortUrl, err := store.SaveUrl(request.LongUrl)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}
	response := config.Response{ShortURL: shortUrl}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/r/")
	fmt.Println("Запрос редиректа для короткого ID:", shortID)

	// Ищем соответствующую длинную ссылку в базе
	longUrl, err := store.GetLongUrl(shortID) // Используем метод GetLongUrl
	if err != nil {
		if err.Error() == "URL не найден" { // Проверяем, что это ошибка "не найдено"
			fmt.Println("Ошибка: короткий ID не найден в базе:", shortID)
			http.NotFound(w, r)
			return
		}
		fmt.Println("Ошибка базы данных:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	fmt.Println("Редирект на:", longUrl)
	http.Redirect(w, r, longUrl, http.StatusFound)

}

func main() {
	var err error
	store, err = postgres.NewUrlStore()
	if err != nil {
		log.Fatalf("Ошибка инициализации PostgreSQL: %v", err)
	}
	defer func(store *postgres.URLStore) {
		err := store.Close()
		if err != nil {

		}
	}(store)

	// Обработчик для статических файлов (например, вашей HTML-страницы)
	fs := http.FileServer(http.Dir("./front"))
	http.Handle("/", fs)

	// Эндпоинт для сокращения URL
	http.HandleFunc("/shorten", shortenHandler)

	// Обработчик для всех остальных путей (редирект по коротким URL)
	http.HandleFunc("/r/", redirectHandler)

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
