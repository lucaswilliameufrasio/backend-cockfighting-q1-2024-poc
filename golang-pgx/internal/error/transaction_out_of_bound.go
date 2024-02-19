package custom_error

type TransactionOutOfBoundError struct{}

func (r *TransactionOutOfBoundError) Error() string {
	return "Falha ao executar transação, pois não há limite disponível"
}
