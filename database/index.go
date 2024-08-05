package database

import (
	_ "database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/fabiofcferreira/xd-rsync/logger"
)

type DatabaseClient struct {
	db     *sqlx.DB
	logger *logger.Logger
}

type DatabaseClientCreationInput struct {
	DSN    string
	Logger *logger.Logger
}

func CreateClient(input *DatabaseClientCreationInput) (*DatabaseClient, error) {
	var err error
	var dbConnection *sqlx.DB

	service := &DatabaseClient{
		logger: input.Logger,
	}

	// Setup database connection
	service.logger.Info("init_db_connection", "Initialising DB connection...", nil)
	dbConnection, err = sqlx.Open("mysql", input.DSN)
	if err != nil {
		service.logger.Error("failed_init_db_connection", "Failed to establish DB connection", &map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	err = dbConnection.Ping()
	if err != nil {
		service.logger.Error("failed_db_ping", "Failed to ping DB", &map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.New("database ping failed")
	}

	service.db = dbConnection

	service.logger.Info("finished_init_db_connection", "Established DB connection successfully", nil)
	return service, nil
}
