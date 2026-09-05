package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	pool, err := initDB()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rdb, err := initRedis()

	if err != nil {
		log.Println(err)
		rdb = nil

	}

	defer pool.Close()

	app := &App{
		DB:    pool,
		Redis: rdb,
	}

	/*
			ctx := context.Background()

				err = app.Redis.Set(
					ctx,
					"url:test",
					"https://google.com",
					time.Hour,
				).Err()

				value, err := app.Redis.Get(
					ctx,
					"url:test",
				).Result()
		fmt.Println(value)
	*/

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", app.shorten)
	mux.HandleFunc("GET /{code}", app.get_code)

	mux.HandleFunc("GET /{$}", homeHandler)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := c.Handler(mux)

	limiter := newLimiter(10, 1)
	wrapped := rateLimitMiddleware(limiter, handler)
	port := os.Getenv("PORT")
	log.Println("Server started at port ", port)
	go startClickWorker(pool)
	log.Fatal(http.ListenAndServe(":"+port, wrapped))
}
