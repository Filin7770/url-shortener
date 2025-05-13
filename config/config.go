package config

import "net/http"

type Request struct {
	LongUrl string `json:"longUrl"`
}

type Response struct {
	ShortURL string `json:"shortUrl"`
}

func GetBaseURL(r *http.Request) string {
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	return scheme + "://" + host
}
