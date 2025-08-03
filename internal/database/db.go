package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

type Version struct {
	ID       string
	FileKey  string
	FileName string
}

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

func CreateVersion(packageID, version, fileKey, fileName string) (string, error) {
	var id string
	err := DB.QueryRow(
		context.Background(),
		"INSERT INTO versions (package_id, version, file_key, file_name) VALUES ($1, $2, $3, $4) RETURNING id",
		packageID, version, fileKey, fileName,
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

func GetVersion(packageID, version string) (Version, error) {
	var v Version
	err := DB.QueryRow(
		context.Background(),
		"SELECT id, file_key, file_name FROM versions WHERE package_id=$1 AND version=$2",
		packageID, version,
	).Scan(&v.ID, &v.FileKey, &v.FileName)
	return v, err
}
