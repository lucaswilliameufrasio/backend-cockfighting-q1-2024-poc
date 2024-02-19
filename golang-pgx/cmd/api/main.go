package main

import (
	"backend-cockfighting-q1-2024-golang-pgx-poc/helper"
	"backend-cockfighting-q1-2024-golang-pgx-poc/internal/customer"
	"backend-cockfighting-q1-2024-golang-pgx-poc/internal/transaction"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultMaxConns = int32(30)
const defaultMinConns = int32(30)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to create a pgxpool config, error: ", err)
		os.Exit(1)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns

	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())

		return nil
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	err = dbPool.Ping(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to ping the databse: %v\n", err)
		os.Exit(1)
	}

	var greeting string
	err = dbPool.QueryRow(ctx, "SELECT 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Printf("Initial query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	mux := http.NewServeMux()

	transactionRepository := transaction.NewTransactionRepository(dbPool)
	transactionService := transaction.NewTransactionService(transactionRepository)
	customerRepository := customer.NewCustomerRepository(dbPool)
	customerService := customer.NewCustomerService(customerRepository, transactionService)
	transactionController := customer.NewCustomerController(customerService)

	mux.HandleFunc("GET /health-check", healthCheck)
	mux.HandleFunc("POST /clientes/{id}/transacoes", transactionController.MakeTransaction)
	mux.HandleFunc("GET /clientes/{id}/extrato", transactionController.LoadStatement)

	address := fmt.Sprintf(":%s", os.Getenv("PORT"))

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	log.Printf("Listening on %s\n", address)
	server.ListenAndServe() // Run the http server
}

var healthCheckResponse = []byte("{\"message\": \"ok\"}")

func healthCheck(w http.ResponseWriter, r *http.Request) {
	helper.MakeHttpResponseFromJSONBytes(w, http.StatusOK, healthCheckResponse)
	return
}
