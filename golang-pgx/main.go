package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	ctx := context.Background()
	defer ctx.Done()
	dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	dbPool.Config().AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())

		return nil
	}
	defer dbPool.Close()

	db = dbPool

	var greeting string
	err = db.QueryRow(ctx, "SELECT 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Printf("Initial query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health-check", healthCheck)
	mux.HandleFunc("POST /clientes/{id}/transacoes", saveTransaction)
	mux.HandleFunc("GET /clientes/{id}/extrato", loadBankStatement)

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
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(healthCheckResponse)
}

type SaveTransactionRequestBody struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

func saveTransaction(w http.ResponseWriter, r *http.Request) {
	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("{\"message\":\"Teus dados tão tudo inválidos, macho\"}"))
		return
	}

	if customerId < 1 || customerId > 5 {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"message\":\"Cliente não encontrado\"}"))
		return
	}

	// log.Printf("rélou %s", customerId)

	var transaction SaveTransactionRequestBody
	err = json.NewDecoder(r.Body).Decode(&transaction)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("{\"message\":\"Teus dados tão tudo inválidos, macho\"}"))
		return
	}

	// log.Printf("Transaction type %s\n", transaction.Type)
	// log.Printf("Transaction description %s\n", transaction.Description)
	// log.Printf("Transaction value %d\n", transaction.Value)

	descriptionLength := len(transaction.Description)

	if (transaction.Type != "c" && transaction.Type != "d") || descriptionLength <= 0 || descriptionLength > 10 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("{\"message\":\"Teus dados tão tudo inválidos, macho\"}"))
		return
	}

	ctx := context.Background()

	defer ctx.Done()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	// tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	defer tx.Rollback(ctx)

	if err != nil {
		// fmt.Fprintf(os.Stderr, "transaction initiation failed: %v\n", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	// var balance int64
	var limit int64

	// err = tx.QueryRow(ctx, "SELECT balance, \"limit\" FROM customers WHERE customers.id = $1 FOR UPDATE;", customerId).Scan(&balance, &limit)
	err = tx.QueryRow(ctx, "SELECT \"limit\" FROM customers WHERE customers.id = $1 FOR UPDATE;", customerId).Scan(&limit)

	if err != nil && err.Error() == "no rows in result set" {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"message\":\"Cliente não encontrado\"}"))
		return
	}

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	// log.Printf("Customer balance on %d\n", balance)
	// log.Printf("Customer limit on %d\n", limit)

	_, err = tx.Exec(ctx, `
  		INSERT INTO
			transactions (
			  description,
			  "type",
			  "value",
			  created_at,
			  customer_id
			)
  		VALUES
			($1, $2, $3, NOW(), $4);
	`, transaction.Description,
		transaction.Type,
		transaction.Value,
		customerId,
	)

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		// fmt.Fprintf(os.Stderr, "Test: %s\n", err.Error())

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Message == "The transaction cannot exceed the bounds of the balance." {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("{\"message\":\"Falha ao executar transação, pois não há limite disponível\"}"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	var updated_balance int64

	if transaction.Type == "d" {
		err = tx.QueryRow(ctx, "UPDATE customers SET balance = balance - $1 WHERE id = $2 RETURNING balance;", transaction.Value, customerId).Scan(&updated_balance)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"message\":\"Deu boró\"}"))
			return
		}
	} else {
		err = tx.QueryRow(ctx, "UPDATE customers SET balance = balance + $1 WHERE id = $2 RETURNING balance;", transaction.Value, customerId).Scan(&updated_balance)

		if err != nil {
			// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"message\":\"Deu boró\"}"))
			return
		}
	}

	err = tx.Commit(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "transaction commit failed: %v\n", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"limite\":%d,\"saldo\":%d}", limit, updated_balance)))
}

type CustomerStatement struct {
	Balance     int       `json:"total"`
	GeneratedAt time.Time `json:"data_extrato"`
	Limit       int       `json:"limite"`
}

type Transaction struct {
	Value       int       `json:"valor"`
	Type        string    `json:"tipo"`
	Description string    `json:"descricao"`
	CreatedAt   time.Time `json:"realizada_em"`
}

type Transactions []Transaction

type CustomerStatementResponse struct {
	Balance     int    `json:"total"`
	GeneratedAt string `json:"data_extrato"`
	Limit       int    `json:"limite"`
}

type TransactionResponse struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
	CreatedAt   string `json:"realizada_em"`
}

type TransactionsResponse []TransactionResponse

type LoadBankStatementResponse struct {
	CustomerStatementResponse `json:"saldo"`
	TransactionsResponse      `json:"ultimas_transacoes"`
}

func loadBankStatement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	defer ctx.Done()

	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("{\"message\":\"Teus dados tão tudo inválidos, macho\"}"))
		return
	}

	// customer := findCachedCustomerById(customerId)

	// if customer == nil {
	// 	// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusNotFound)
	// 	w.Write([]byte("{\"message\":\"Cliente não encontrado\"}"))
	// 	return
	// }

	if customerId < 1 || customerId > 5 {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"message\":\"Cliente não encontrado\"}"))
		return
	}

	var customerStatement CustomerStatement

	// customerStatement.Limit = customer.Limit

	err = db.QueryRow(ctx, "SELECT balance, \"limit\", NOW() FROM customers WHERE customers.id = $1;", customerId).Scan(&customerStatement.Balance, &customerStatement.Limit, &customerStatement.GeneratedAt)
	// err = db.QueryRow(ctx, "SELECT balance, NOW() FROM customers WHERE customers.id = $1;", customerId).Scan(&customerStatement.Balance, &customerStatement.GeneratedAt)

	if err != nil && err.Error() == "no rows in result set" {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"message\":\"Cliente não encontrado\"}"))
		return
	}

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	transactions, err := loadLastTenTransactions(ctx, customerId)

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\":\"Deu boró\"}"))
		return
	}

	transactionsResponse := make([]TransactionResponse, 0, 10)

	for _, transaction := range transactions {
		transactionResponse := TransactionResponse{
			Type:        transaction.Type,
			Value:       transaction.Value,
			Description: transaction.Description,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		}
		transactionsResponse = append(transactionsResponse, transactionResponse)
	}

	response, err := json.Marshal(LoadBankStatementResponse{
		CustomerStatementResponse: CustomerStatementResponse{
			Balance:     customerStatement.Balance,
			GeneratedAt: customerStatement.GeneratedAt.Format(time.RFC3339),
			Limit:       customerStatement.Limit,
		},
		TransactionsResponse: transactionsResponse,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func loadLastTenTransactions(ctx context.Context, customerId int) (Transactions, error) {
	rows, err := db.Query(ctx, "SELECT description, type, value, created_at FROM transactions WHERE customer_id = $1 ORDER BY transactions.id DESC LIMIT 10", customerId)

	if err != nil && err.Error() == "no rows in result set" {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return Transactions{}, nil
	}

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return Transactions{}, err
	}

	transactions := make(Transactions, 0, 10)

	for rows.Next() {
		var transaction Transaction
		_ = rows.Scan(&transaction.Description, &transaction.Type, &transaction.Value, &transaction.CreatedAt)
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
