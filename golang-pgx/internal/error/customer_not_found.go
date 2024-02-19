package custom_error

type CustomerNotFoundError struct{}

func (r *CustomerNotFoundError) Error() string {
	return "Cliente não encontrado"
}
