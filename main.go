package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB *pgxpool.Pool
}

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encode(n uint64) string {
	if n == 0 {
		return "0"
	}

	var out []byte

	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j+1 {
		out[i], out[j] = out[j], out[i]
	}

	return string(out)
}

func decode(str string) uint64 {
	var out int = 0
	for _, s := range str {
		pos := strings.IndexRune(alphabet, s)
		if pos == -1 {
			panic("Invalid Character")
		}

		out = out*62 + pos
	}

	return uint64(out)

}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed for this endpoint", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var id uint64
	err := a.DB.QueryRow(context.Background(), `SELECT nextval('urls_id_seq')`).Scan(&id)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := encode(id)

	_, err = a.DB.Exec(context.Background(), `INSERT INTO urls (id, code, original_url) VALUES ($1, $2, $3)`, id, code, req.URL)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(ShortenResponse{
		ShortURL: code,
	})
}

func (a *App) get_code(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	id := decode(code)

	var original_url string
	err := a.DB.QueryRow(context.Background(), `SELECT original_url FROM url WHERE id = $1`, id).Scan(&original_url)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.Redirect(w, r, original_url, http.StatusFound)

}

// func redirect() {
//
// }

func main() {

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	defer pool.Close()

	app := &App{
		DB: pool,
	}
	http.HandleFunc("/shorten", app.shorten)
	http.HandleFunc("/", app.get_code)
	fmt.Println("Server started at port 5000")
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)

}
