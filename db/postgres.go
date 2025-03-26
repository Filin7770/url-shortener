package postgres

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

type URLStore struct {
	db *sql.DB
}

func NewUrlStore() (*URLStore, error) {
	connStr := "postgres://admin:1234@localhost:5432/url_shortener?sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка ping PostgreSQL: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
	    id serial PRIMARY KEY,
	    short_url VARCHAR(6) UNIQUE NOT NULL,
	    long_url TEXT UNIQUE NOT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
`)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания таблицы: %v", err)
	}
	return &URLStore{db: db}, nil
}

func generateShortUrl(longUrl string) (string, error) {
	hash := sha256.Sum256([]byte(longUrl))
	hexHash := hex.EncodeToString(hash[:])
	shortHash := hexHash[:6]
	return shortHash, nil
}
func (store *URLStore) SaveUrl(longUrl string) (string, error) {
	var existingUrl string
	err := store.db.QueryRow(`
	SELECT short_url 
	FROM urls 
	WHERE long_url = $1`, longUrl).Scan(&existingUrl)

	if err == nil {
		return "http://localhost:8080/r/" + existingUrl, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("ошибка проверки существующей ссылки: %v", err)
	}

	shortUrl, err := generateShortUrl(longUrl)
	if err != nil {
		return "", err
	}

	_, err = store.db.Exec(`
		INSERT INTO urls (short_url, long_url)
		VALUES ($1, $2)
		ON CONFLICT (long_url) 
		DO UPDATE SET short_url = EXCLUDED.short_url`,
		shortUrl, longUrl)

	if err != nil {
		return "", fmt.Errorf("ошибка сохранения URL: %v", err)
	}

	return "http://localhost:8080/r/" + shortUrl, nil
}

func (store *URLStore) GetLongUrl(shortUrl string) (string, error) {
	var longUrl string
	err := store.db.QueryRow(`
		SELECT long_url 
		FROM urls 
		WHERE short_url = $1`, shortUrl).Scan(&longUrl)

	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("URL не найден")
	}
	if err != nil {
		return "", fmt.Errorf("ошибка получения URL: %v", err)
	}
	return longUrl, nil
}

func (store *URLStore) Close() error {
	return store.db.Close()
}
