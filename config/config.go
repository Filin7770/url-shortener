package config

type Request struct {
	LongUrl string `json:"longUrl"`
}

type Response struct {
	ShortURL string `json:"shortUrl"`
}
