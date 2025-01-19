package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	dbService *DBService
)

type DBService struct {
	DB *sql.DB
}

func Service() *DBService {
	if dbService == nil {
		dbService = &DBService{}
	}
	return dbService
}

func (d *DBService) Initialize() error {
	dbConn, err := sql.Open("pgx", config.GetDBString())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	d.DB = dbConn
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = dbConn.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("couldn't connect to db: %w", err)
	}

	return nil
}
