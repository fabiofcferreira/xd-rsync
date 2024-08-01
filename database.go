package xd_rsync

import "github.com/fabiofcferreira/xd-rsync/logger"

type ServiceInitialisationInput struct {
	DSN    string
	Logger *logger.Logger
}

type DatabaseService interface {
	Init(input *ServiceInitialisationInput) error

	GetProductByReferece(id string) (XdProduct, error)
}
