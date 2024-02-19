package transaction

import (
	custom_error "backend-cockfighting-q1-2024-golang-pgx-poc/internal/error"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return TransactionRepository{
		db: db,
	}
}

func (tctx TransactionRepository) SaveTransaction(ctx context.Context, input MakeTransactionInput) (MakeTransactionOutput, error) {
	tx, err := tctx.db.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return MakeTransactionOutput{}, err
	}

	var limit int
	err = tx.QueryRow(ctx, "SELECT \"limit\" FROM customers WHERE customers.id = $1 FOR UPDATE;", input.CustomerId).Scan(&limit)

	if err != nil && err.Error() == "no rows in result set" {
		return MakeTransactionOutput{}, &custom_error.CustomerNotFoundError{}
	}

	if err != nil {
		return MakeTransactionOutput{}, err
	}

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
	`, input.Description,
		input.Type,
		input.Value,
		input.CustomerId,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Message == "The transaction cannot exceed the bounds of the balance." {
			return MakeTransactionOutput{}, &custom_error.TransactionOutOfBoundError{}
		}

		return MakeTransactionOutput{}, err
	}

	var updated_balance int

	if input.Type == "d" {
		err = tx.QueryRow(ctx, "UPDATE customers SET balance = balance - $1 WHERE id = $2 RETURNING balance;", input.Value, input.CustomerId).Scan(&updated_balance)
		if err != nil {
			return MakeTransactionOutput{}, err
		}
	} else {
		err = tx.QueryRow(ctx, "UPDATE customers SET balance = balance + $1 WHERE id = $2 RETURNING balance;", input.Value, input.CustomerId).Scan(&updated_balance)

		if err != nil {
			return MakeTransactionOutput{}, err
		}
	}

	err = tx.Commit(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "transaction commit failed: %v\n", err)
		return MakeTransactionOutput{}, err
	}

	return MakeTransactionOutput{
		Balance: updated_balance,
		Limit:   limit,
	}, nil
}

func (tctx TransactionRepository) LoadLastTenTransactions(ctx context.Context, customerId int) ([]Transaction, error) {
	rows, err := tctx.db.Query(ctx, "SELECT description, type, value, created_at FROM transactions WHERE customer_id = $1 ORDER BY transactions.id DESC LIMIT 10", customerId)

	defer rows.Close()

	if err != nil && err.Error() == "no rows in result set" {
		return Transactions{}, nil
	}

	if err != nil {
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
