package vault

import (
	"net/http"

	"go.uber.org/zap"
)

type Vault struct {
	address string
	token   string
	logger  *zap.Logger
	client  *http.Client
}

func New(logger *zap.Logger, address, token string) *Vault {
	return &Vault{
		logger:  logger,
		address: address,
		token:   token,
		client:  &http.Client{},
	}
}
