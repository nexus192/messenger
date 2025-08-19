package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"messenger/src/internal/data/model"

	_ "github.com/jackc/pgx/v5/stdlib"

	"golang.org/x/crypto/bcrypt"
)

type PostgresRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewPGRepository(dsn string, log *slog.Logger) (*PostgresRepo, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db, log: log}, nil
}

func (r *PostgresRepo) CreateUser(ctx context.Context, nick, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO users (nick, password_hash) VALUES ($1, $2)",
		nick, string(hash),
	)
	if err != nil {
		r.log.Error("CreateUser failed", "error", err)
	}
	return err
}

func (r *PostgresRepo) CheckUser(ctx context.Context, nick, password string) (int, error) {
	var id int
	var hash string

	err := r.db.QueryRowContext(ctx,
		"SELECT id, password_hash FROM users WHERE nick=$1",
		nick,
	).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("user not found")
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return 0, errors.New("invalid password")
	}

	return id, nil
}

func (r *PostgresRepo) SaveMessage(userID int, message string) error {
	_, err := r.db.Exec(
		"INSERT INTO history (user_id, message, created_at) VALUES ($1, $2, $3)",
		userID, message, time.Now(),
	)
	if err != nil {
		r.log.Error("SaveMessage failed", "error", err)
	}
	return err
}

func (r *PostgresRepo) GetHistory(userID int) ([]model.Message, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, message, created_at FROM history WHERE user_id=$1 ORDER BY created_at",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
