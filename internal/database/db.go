package database

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func CreateVersion(packageID, version, fileKey, fileName string, size int64) (string, error) {
	var id string
	err := DB.QueryRow(
		context.Background(),
		`INSERT INTO versions (package_id, version, file_key, file_name, size)
         VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		packageID, version, fileKey, fileName, size,
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

type Package struct {
	ID      string
	Name    string
	OwnerID string
}

func ListPackages() ([]Package, error) {
	rows, err := DB.Query(context.Background(),
		"SELECT id, name, owner_id FROM packages ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []Package
	for rows.Next() {
		var p Package
		if err := rows.Scan(&p.ID, &p.Name, &p.OwnerID); err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}

type VersionInfo struct {
	ID        string
	Version   string
	FileName  string
	Size      int64
	Downloads int64
	CreatedAt time.Time
}

func ListVersions(packageID string) ([]VersionInfo, error) {
	rows, err := DB.Query(context.Background(),
		`SELECT id, version, file_name, size, downloads, created_at
         FROM versions
         WHERE package_id=$1
         ORDER BY created_at DESC`,
		packageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []VersionInfo
	for rows.Next() {
		var v VersionInfo
		if err := rows.Scan(&v.ID, &v.Version, &v.FileName, &v.Size, &v.Downloads, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func DeleteVersion(packageID, version string) error {
	_, err := DB.Exec(
		context.Background(),
		"DELETE FROM versions WHERE package_id=$1 AND version=$2",
		packageID, version,
	)
	return err
}

func IncrementDownloadCount(packageID, version string) error {
	_, err := DB.Exec(context.Background(),
		`UPDATE versions SET downloads = downloads + 1 WHERE package_id=$1 AND version=$2`,
		packageID, version,
	)
	return err
}

func SearchPackages(query string, limit, offset int) ([]Package, error) {
	rows, err := DB.Query(context.Background(),
		`SELECT id, name, owner_id
         FROM packages
         WHERE name ILIKE '%' || $1 || '%'
         ORDER BY name ASC
         LIMIT $2 OFFSET $3`,
		query, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []Package
	for rows.Next() {
		var p Package
		if err := rows.Scan(&p.ID, &p.Name, &p.OwnerID); err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}
