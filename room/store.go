package room

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrewbenington/go-spotify/db/gen"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) All(ctx context.Context) ([]Room, error) {
	rows, err := gen.New(s.db).AllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all rooms: %w", err)
	}
	rooms := make([]Room, len(rows))
	for i, row := range rows {
		rooms[i] = roomFromRow(row)
	}
	return rooms, nil
}

func (s *Store) GetByCode(ctx context.Context, code string) (Room, error) {
	row, err := gen.New(s.db).FindRoomByCode(ctx, code)
	if err != nil {
		return Room{}, fmt.Errorf("get room by code: %w", err)
	}
	return roomFromRow(row), nil
}

func (s *Store) Insert(ctx context.Context, name string) (Room, error) {
	row, err := gen.New(s.db).InsertRoom(ctx, name)
	if err != nil {
		return Room{}, fmt.Errorf("insert room: %w", err)
	}
	return roomFromRow(row), nil
}

func roomFromRow(row gen.Room) Room {
	return Room{
		ID:      row.ID.String(),
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created.Time,
	}
}
