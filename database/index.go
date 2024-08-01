package database

import (
	_ "database/sql"
	"errors"

	"github.com/fabiofcferreira/xd-rsync/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db     *sqlx.DB
	logger *logger.Logger
}

type ServiceInitialisationInput struct {
	DSN    string
	Logger *logger.Logger
}

func (s *Service) Init(input *ServiceInitialisationInput) error {
	var err error
	var dbConnection *sqlx.DB

	// Setup logger
	s.logger = input.Logger

	// Setup database connection
	s.logger.Info("init_db_connection", "Initialising DB connection...", nil)
	dbConnection, err = sqlx.Open("mysql", input.DSN)
	if err != nil {
		s.logger.Error("failed_init_db_connection", "Failed to establish DB connection", &map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	err = dbConnection.Ping()
	if err != nil {
		s.logger.Error("failed_db_ping", "Failed to ping DB", &map[string]interface{}{
			"error": err.Error(),
		})
		return errors.New("database ping failed")
	}

	s.logger.Info("finished_init_db_connection", "Established DB connection successfully", nil)

	s.db = dbConnection
	return nil
}
