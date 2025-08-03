package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(databaseURL string) error {
	var err error
	DB, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return nil
}

func CreateUser(email, passwordHash string) (string, error) {
	var id string
	err := DB.QueryRow(
		context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash,
	).Scan(&id)
	return id, err
}

func GetUserByEmail(email string) (string, string, error) {
	var id, passwordHash string
	err := DB.QueryRow(
		context.Background(),
		"SELECT id, password_hash FROM users WHERE email=$1",
		email,
	).Scan(&id, &passwordHash)
	return id, passwordHash, err
}

func CreatePackage(name, ownerID string) (string, error) {
	var id string
	err := DB.QueryRow(
		context.Background(),
		"INSERT INTO packages (name, owner_id) VALUES ($1, $2) RETURNING id",
		name, ownerID,
	).Scan(&id)
	return id, err
}
