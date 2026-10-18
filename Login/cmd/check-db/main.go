package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	connection := os.Getenv("DATABASE_URL")

	if connection == "" {
		file, err := os.Open(".env")
		if err != nil {
			fmt.Println("Login folder থেকে চালাও; .env পাওয়া যায়নি।")
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			key, value, ok := strings.Cut(line, "=")
			if ok && strings.TrimSpace(key) == "DATABASE_URL" {
				connection = strings.Trim(
					strings.TrimSpace(value), "\"'",
				)
			}
		}
		if scanner.Err() != nil {
			fmt.Println(".env পড়তে সমস্যা হয়েছে।")
			os.Exit(1)
		}
	}

	if connection == "" {
		fmt.Println("In .env- DATABASE_URL is not found ।")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", connection)
	if err != nil {
		fmt.Println("Connection string format is not okay।")
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
	message := strings.ReplaceAll(err.Error(), connection, "[DATABASE_URL hidden]")
	fmt.Println("Connection failed:", message)
	os.Exit(1)
}  
	fmt.Println("Supabase connection successful!")
}