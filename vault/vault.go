package vault

import "go.uber.org/zap"

type Vault struct {
	Address string
	Token   string
	Logger  *zap.Logger
}
