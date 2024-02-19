package customer

import (
	custom_error "backend-cockfighting-q1-2024-golang-pgx-poc/internal/error"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) CustomerRepository {
	return CustomerRepository{
		db: db,
	}
}

func (tctx CustomerRepository) FindStatusById(ctx context.Context, customerId int) (FindCustomerStatusByIdOutput, error) {
	var result FindCustomerStatusByIdOutput

	err := tctx.db.QueryRow(ctx, "SELECT balance, \"limit\" FROM customers WHERE customers.id = $1;", customerId).Scan(&result.Balance, &result.Limit)

	if err != nil && err.Error() == "no rows in result set" {
		return result, &custom_error.CustomerNotFoundError{}
	}

	if err != nil {
		return result, err
	}

	return result, nil
}
