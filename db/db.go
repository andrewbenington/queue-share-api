package db

import (
	"database/sql"
	"fmt"

	"github.com/andrewbenington/go-spotify/config"
	"github.com/andrewbenington/go-spotify/room"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	dbService *DBService
)

type DBService struct {
	DB        *sql.DB
	RoomStore *room.Store
}

func Service() *DBService {
	if dbService == nil {
		dbService = &DBService{}
	}
	return dbService
}

func (d *DBService) Initialize() error {
	cfg := config.GetConfig()
	dbConn, err := sql.Open("pgx", cfg.GetDBString())
	if err != nil {
		return fmt.Errorf("connect to database: %w\n", err)
	}
	d.DB = dbConn

	d.initStores()
	return nil
}

func (d *DBService) initStores() {
	d.RoomStore = room.NewStore(d.DB)
}
