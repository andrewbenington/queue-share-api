package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbService *DBService
)

type DBService struct {
	Pool          *pgxpool.Pool
	usedConnCount int
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

	cfg, err := pgxpool.ParseConfig(config.GetDBString())
	if err != nil {
		return fmt.Errorf("get db pool config: %w", err)
	}
	cfg.MaxConns = 30
	cfg.MaxConnIdleTime = time.Minute * 5
	cfg.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		d.usedConnCount++
		log.Printf("Pool usage: %d/30", d.usedConnCount)
		return true
	}

	cfg.AfterRelease = func(c *pgx.Conn) bool {
		d.usedConnCount--
		log.Printf("Pool usage: %d/30", d.usedConnCount)
		return true
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
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
