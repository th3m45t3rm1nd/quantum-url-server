package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func initDB() (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	return pool, err
}

func initRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	ctx := context.Background()
	err := rdb.Ping(ctx).Err()

	return rdb, err
}
