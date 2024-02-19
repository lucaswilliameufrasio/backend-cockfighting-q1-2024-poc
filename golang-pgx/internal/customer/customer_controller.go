package customer

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

type CustomerController struct {
	customerService CustomerService
}

func NewCustomerController(customerService CustomerService) CustomerController {
	return CustomerController{
		customerService: customerService,
	}
}

type MakeTransactionRequestBody struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

func (tctx CustomerController) MakeTransaction(w http.ResponseWriter, r *http.Request) {
	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	if customerId < 1 || customerId > 5 {
		helper.MakeHttpNotFoundErrorResponse(w, "Cliente não encontrado")
		return
	}

	var transactionInput MakeTransactionRequestBody
	err = json.NewDecoder(r.Body).Decode(&transactionInput)

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	descriptionLength := len(transactionInput.Description)

	if (transactionInput.Type != "c" && transactionInput.Type != "d") || descriptionLength <= 0 || descriptionLength > 10 {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	ctx := context.Background()

	defer ctx.Done()

	customerStatement, err := tctx.customerService.MakeTransaction(ctx, MakeTransactionInput{
		CustomerId:  customerId,
		Description: transactionInput.Description,
		Type:        transactionInput.Type,
		Value:       transactionInput.Value,
	})

	if err != nil && errors.Is(err, &custom_error.CustomerNotFoundError{}) {
		helper.MakeHttpNotFoundErrorResponse(w, err.Error())
		return
	}

	if err != nil && errors.Is(err, &custom_error.TransactionOutOfBoundError{}) {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, err.Error())
		return
	}

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	helper.MakeHttpResponseFromJSONString(w, http.StatusOK, fmt.Sprintf("{\"limite\":%d,\"saldo\":%d}", customerStatement.Limit, customerStatement.Balance))
	return
}

type CustomerStatusResponse struct {
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
	Status       CustomerStatusResponse `json:"saldo"`
	Transactions TransactionsResponse   `json:"ultimas_transacoes"`
}

func (tctx CustomerController) LoadStatement(w http.ResponseWriter, r *http.Request) {
	customerId, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		helper.MakeHttpUnprocessableEntityErrorResponse(w, "Não foi possível processar sua requisição, pois foram enviados dados inválidos")
		return
	}

	if customerId < 1 || customerId > 5 {
		customerNotFoundError := &custom_error.CustomerNotFoundError{}
		helper.MakeHttpNotFoundErrorResponse(w, customerNotFoundError.Error())
		return
	}

	ctx := context.Background()
	defer ctx.Done()

	statementResult, err := tctx.customerService.LoadStatement(ctx, customerId)

	if err != nil && errors.Is(err, &custom_error.CustomerNotFoundError{}) {
		helper.MakeHttpNotFoundErrorResponse(w, err.Error())
		return
	}

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	transactionsResponse := make([]TransactionResponse, 0, 10)

	for _, transaction := range statementResult.Transactions {
		transactionResponse := TransactionResponse{
			Type:        transaction.Type,
			Value:       transaction.Value,
			Description: transaction.Description,
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		transactionsResponse = append(transactionsResponse, transactionResponse)
	}

	response, err := json.Marshal(LoadBankStatementResponse{
		Status: CustomerStatusResponse{
			Balance:     statementResult.Status.Balance,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Limit:       statementResult.Status.Limit,
		},
		Transactions: transactionsResponse,
	})

	if err != nil {
		helper.MakeHttpInternalServerErrorResponse(w)
		return
	}

	helper.MakeHttpResponseFromJSONBytes(w, http.StatusOK, response)
	return
}
