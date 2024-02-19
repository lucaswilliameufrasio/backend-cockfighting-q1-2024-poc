package transaction

import (
	"backend-cockfighting-q1-2024-golang-pgx-poc/helper"
	custom_error "backend-cockfighting-q1-2024-golang-pgx-poc/internal/error"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type TransactionController struct {
	transactionRepository TransactionRepository
}

func NewTransactionController(transactionRepository TransactionRepository) TransactionController {
	return TransactionController{
		transactionRepository: transactionRepository,
	}
}

type SaveTransactionRequestBody struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

type SaveTransactionInput struct {
	CustomerId  int
	Value       int
	Type        string
	Description string
}

func (tctx TransactionController) SaveTransaction(w http.ResponseWriter, r *http.Request) {
	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	if customerId < 1 || customerId > 5 {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		helper.MakeHttpNotFoundErrorResponse(w, "Cliente não encontrado")
		return
	}

	// log.Printf("rélou %s", customerId)

	var transaction SaveTransactionRequestBody
	err = json.NewDecoder(r.Body).Decode(&transaction)

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	// log.Printf("Transaction type %s\n", transaction.Type)
	// log.Printf("Transaction description %s\n", transaction.Description)
	// log.Printf("Transaction value %d\n", transaction.Value)

	descriptionLength := len(transaction.Description)

	if (transaction.Type != "c" && transaction.Type != "d") || descriptionLength <= 0 || descriptionLength > 10 {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	ctx := context.Background()

	defer ctx.Done()

	customerStatement, err := tctx.transactionRepository.SaveTransaction(ctx, SaveTransactionInput{
		CustomerId:  customerId,
		Description: transaction.Description,
		Type:        transaction.Type,
		Value:       transaction.Value,
	})

	if err != nil && errors.Is(err, &custom_error.CustomerNotFoundError{}) {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		helper.MakeHttpNotFoundErrorResponse(w, err.Error())
		return
	}

	if err != nil && errors.Is(err, &custom_error.TransactionOutOfBoundError{}) {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		helper.MakeHttpUnprocessableEntityErrorResponse(w, err.Error())
		return
	}

	if err != nil {
		// fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)

		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	helper.MakeHttpResponseFromJSONString(w, http.StatusOK, fmt.Sprintf("{\"limite\":%d,\"saldo\":%d}", customerStatement.Limit, customerStatement.Balance))
	return
}

type CustomerStatement struct {
	Balance int `json:"total"`
	Limit   int `json:"limite"`
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

func (tctx TransactionController) LoadBankStatement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	defer ctx.Done()

	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	if customerId < 1 || customerId > 5 {
		helper.MakeHttpNotFoundErrorResponse(w, "Cliente não encontrado")
		return
	}

	customerStatement, err := tctx.transactionRepository.LoadCustomerStatement(ctx, customerId)

	if err != nil && errors.Is(err, &custom_error.CustomerNotFoundError{}) {
		helper.MakeHttpNotFoundErrorResponse(w, err.Error())
		return
	}

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	transactions, err := tctx.transactionRepository.LoadLastTenTransactions(ctx, customerId)

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	transactionsResponse := make([]TransactionResponse, 0, 10)

	for _, transaction := range transactions {
		transactionResponse := TransactionResponse{
			Type:        transaction.Type,
			Value:       transaction.Value,
			Description: transaction.Description,
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		transactionsResponse = append(transactionsResponse, transactionResponse)
	}

	response, err := json.Marshal(LoadBankStatementResponse{
		CustomerStatementResponse: CustomerStatementResponse{
			Balance:     customerStatement.Balance,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Limit:       customerStatement.Limit,
		},
		TransactionsResponse: transactionsResponse,
	})

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	helper.MakeHttpResponseFromJSONBytes(w, http.StatusOK, response)
	return
}
