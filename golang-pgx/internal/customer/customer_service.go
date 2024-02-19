package customer

import (
	"backend-cockfighting-q1-2024-golang-pgx-poc/internal/transaction"
	"context"
)

type CustomerService struct {
	customerRepository CustomerRepository
	transactionService transaction.TransactionService
}

func NewCustomerService(customerRepository CustomerRepository, transactionService transaction.TransactionService) CustomerService {
	return CustomerService{
		customerRepository: customerRepository,
		transactionService: transactionService,
	}
}

type MakeTransactionInput struct {
	CustomerId  int
	Value       int
	Type        string
	Description string
}

type SaveTransactionOutput struct {
	Balance int
	Limit   int
}

func (tctx CustomerService) MakeTransaction(ctx context.Context, input MakeTransactionInput) (SaveTransactionOutput, error) {
	result, err := tctx.transactionService.MakeTransaction(ctx, transaction.MakeTransactionInput{
		CustomerId:  input.CustomerId,
		Description: input.Description,
		Type:        input.Type,
		Value:       input.Value,
	})

	return SaveTransactionOutput{
		Balance: result.Balance,
		Limit:   result.Limit,
	}, err
}

type FindCustomerStatusByIdOutput struct {
	Balance int
	Limit   int
}

type LoadStatementOutput struct {
	Status       FindCustomerStatusByIdOutput
	Transactions transaction.Transactions
}

func (tctx CustomerService) LoadStatement(ctx context.Context, customerId int) (LoadStatementOutput, error) {
	status, err := tctx.customerRepository.FindStatusById(ctx, customerId)

	if err != nil {
		return LoadStatementOutput{}, err
	}

	transactions, err := tctx.transactionService.LoadLastTenTransactions(ctx, customerId)

	if err != nil {
		return LoadStatementOutput{}, err
	}

	return LoadStatementOutput{
		Status:       status,
		Transactions: transactions,
	}, nil
}
