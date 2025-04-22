package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jxskiss/base62"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UUID_LENGTH_SHORT_CODE = 8

type URLRepository interface {
	Create(ctx context.Context, originalURL string) (string, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Create(ctx context.Context, originalURL string) (string, error) {
	var shortCode string
	shortCode = generateUUIDCode(UUID_LENGTH_SHORT_CODE)
	isCodeUnique, err := r.isCodeUnique(ctx, shortCode)
	for !isCodeUnique {
		shortCode = generateUUIDCode(UUID_LENGTH_SHORT_CODE)
		isCodeUnique, err = r.isCodeUnique(ctx, shortCode)
	}

	const query = `
		INSERT INTO urls(original_url, short_code) 
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO UPDATE SET created_at = EXCLUDED.created_at
		RETURNING short_code
	`

	err = r.pool.QueryRow(ctx, query, originalURL, shortCode).Scan(&shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to create short URL: %w", err)
	}

	return shortCode, nil
}

func (r *PostgresRepo) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	var originalURL string
	err := r.pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get original URL: %w", err)
	}

	return originalURL, nil
}

func (r *PostgresRepo) isCodeUnique(ctx context.Context, shortCode string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)",
		shortCode,
	).Scan(&exists)
	return !exists, err
}

func generateUUIDCode(length int) string {
	uuidStr := uuid.New().String()
	encoded := base62.Encode([]byte(uuidStr))
	return string(encoded[:length])
}

var ErrNotFound = fmt.Errorf("not found")
