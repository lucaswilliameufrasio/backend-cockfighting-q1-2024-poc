package transaction

import (
	"context"
	"time"
)

type TransactionService struct {
	transactionRepository TransactionRepository
}

func NewTransactionService(transactionRepository TransactionRepository) TransactionService {
	return TransactionService{
		transactionRepository: transactionRepository,
	}
}

type MakeTransactionInput struct {
	CustomerId  int
	Value       int
	Type        string
	Description string
}

type MakeTransactionOutput struct {
	Balance int
	Limit   int
}

func (tctx TransactionService) MakeTransaction(ctx context.Context, input MakeTransactionInput) (MakeTransactionOutput, error) {
	return tctx.transactionRepository.SaveTransaction(ctx, input)
}

type Transaction struct {
	Value       int
	Type        string
	Description string
	CreatedAt   time.Time
}

type Transactions []Transaction

func (tctx TransactionService) LoadLastTenTransactions(ctx context.Context, customerId int) (Transactions, error) {
	return tctx.transactionRepository.LoadLastTenTransactions(ctx, customerId)
}
