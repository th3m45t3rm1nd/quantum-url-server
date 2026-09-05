package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func rateLimitMiddleware(l *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}
		log.Println("RemoteAddr:", r.RemoteAddr)
		log.Println("Client IP:", clientIP(r))
		if !l.allow(clientIP(r)) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})

}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	URL, err := validateURL(req.URL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error validating the URL", http.StatusBadRequest)
		return
	}

	var code string

	if req.Alias != "" {
		alias, err := validateAlias(req.Alias)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error with Alias", http.StatusBadRequest)
			return
		}

		_, err = a.DB.Exec(context.Background(), `INSERT INTO urls (code, original_url) VALUES ($1, $2)`, alias, URL)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to create an alias", http.StatusInternalServerError)
			return
		}

		code = alias

	} else {
		var id uint64
		err := a.DB.QueryRow(context.Background(), `SELECT nextval('urls_id_seq')`).Scan(&id)

		if err != nil {
			fmt.Println(err)
		}

		code = encode(id)

		_, err = a.DB.Exec(context.Background(), `INSERT INTO urls (id, code, original_url) VALUES ($1, $2, $3)`, id, code, URL)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
			return

		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(ShortenResponse{
		ShortURL: code,
	})
}

func (a *App) get_code(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	cacheKey := "url:" + code

	if a.Redis != nil {
		cachedURL, err := a.Redis.Get(context.Background(), cacheKey).Result()
		if err == nil {
			select {
			case clickChan <- clickEvent{Code: code, Timestamp: time.Now(), UserAgent: r.UserAgent(), Referrer: r.Referer()}:
			default:
				log.Println("click event dropped - channel full")
			}
			http.Redirect(w, r, cachedURL, http.StatusFound)
			log.Println("CACHE HIT")
			return
		}
	}
	log.Println("CACHE MISS", cacheKey)
	log.Println("DB LOOKUP:", cacheKey)

	var original_url string
	err := a.DB.QueryRow(context.Background(), `SELECT original_url FROM urls WHERE code = $1`, code).Scan(&original_url)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid url"})
	} else {
		err = a.Redis.Set(
			context.Background(),
			cacheKey,
			original_url,
			time.Hour,
		).Err()
		log.Println("CACHE SET:", cacheKey)
		select {
		case clickChan <- clickEvent{Code: code, Timestamp: time.Now(), UserAgent: r.UserAgent(), Referrer: r.Referer()}:
		default:
			log.Println("click event dropped - channel full")
		}
		http.Redirect(w, r, original_url, http.StatusFound)
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "url shortener — th3m45t3rm1nd")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "POST /shorten  {\"url\": \"https://example.com\"}")
	fmt.Fprintln(w, "GET  /{code}   302 redirect")
}
