package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/config"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/andrewbenington/queue-share-api/user"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	dbService *DBService
)

type DBService struct {
	DB        *sql.DB
	RoomStore *room.Store
	UserStore *user.Store
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
		return fmt.Errorf("open database: %w\n", err)
	}
	d.DB = dbConn
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = dbConn.PingContext(ctx)
	if err != nil {
		log.Fatalf("couldn't connect to db: %s", err)
	}

	d.initStores()
	return nil
}

func (d *DBService) initStores() {
	d.RoomStore = room.NewStore(d.DB)
	d.UserStore = user.NewStore(d.DB)
}
