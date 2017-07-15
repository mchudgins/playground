package reverseProxy

type transaction struct{}

func New() (*transaction, error) {
	return &transaction{}, nil
}

func (t *transaction) Create() (int, error) {
	return 0, nil
}
