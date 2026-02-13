package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"warranty_days/internal/config"
	"warranty_days/internal/db"
)

func main() {
	_ = godotenv.Load() // в проде обычно не нужно, там env уже задан

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config error: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("db connection error: ", err)
	}
	defer pool.Close()

	// тест: узнать текущую версию Postgres
	var version string
	if err := pool.QueryRow(context.Background(), "select version()").Scan(&version); err != nil {
		log.Fatal("query error: ", err)
	}
	fmt.Println("Connected to Postgres:", version)

	// На следующем этапе здесь будет запуск HTTP сервера.
}