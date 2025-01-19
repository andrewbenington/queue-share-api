package db

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbService *DBService
)

type DBService struct {
	Pool *pgxpool.Pool
}

func Service() *DBService {
	if dbService == nil {
		dbService = &DBService{}
	}
	return dbService
}

func (d *DBService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, config.GetDBString())
	if err != nil {
		return fmt.Errorf("create connection pool: %w", err)
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("couldn't connect to db: %w", err)
	}

	d.Pool = dbpool

	return nil
}

func (d *DBService) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return d.Pool.BeginTx(ctx, pgx.TxOptions{})
}
