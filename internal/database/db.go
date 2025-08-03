package database

import (
	"context"
	"fmt"
	"strings"

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

func CreateVersion(packageID, version, fileKey string) (string, error) {
	var id string
	err := DB.QueryRow(
		context.Background(),
		"INSERT INTO versions (package_id, version, file_key) VALUES ($1, $2, $3) RETURNING id",
		packageID, version, fileKey,
	).Scan(&id)
	return id, err
}

func GetPackageByName(name string) (string, string, error) {
	var id, ownerID string
	err := DB.QueryRow(
		context.Background(),
		"SELECT id, owner_id FROM packages WHERE name=$1",
		strings.ToLower(name),
	).Scan(&id, &ownerID)
	return id, ownerID, err
}

func GetVersion(packageID, version string) (string, string, error) {
	var id, fileKey string
	err := DB.QueryRow(
		context.Background(),
		"SELECT id, file_key FROM versions WHERE package_id=$1 AND version=$2",
		packageID, version,
	).Scan(&id, &fileKey)
	return id, fileKey, err
}
