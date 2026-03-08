package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/go-sql-driver/mysql"
)

var ErrUserNotFound = errors.New("user not found")
var ErrDuplicateUsername = errors.New("duplicate username")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {

	repo := UserRepository{
		db: db,
	}

	_, err := repo.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY AUTO_INCREMENT, 
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at DATETIME
	);`)
	if err != nil {
		slog.Error("error creating users table", "err", err)
	}

	return &repo
}

func (u *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	row := u.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username)

	var user entity.User

	if err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserRepository) Create(ctx context.Context, username string, passwordHash string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, NOW())")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, username, passwordHash)
	if err != nil {
		mySqlErr, ok := err.(*mysql.MySQLError)
		if ok && mySqlErr.Number == 1062 { // 1062 is the error code for duplicate entry
			return 0, ErrDuplicateUsername
		}
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(lastId), nil
}
