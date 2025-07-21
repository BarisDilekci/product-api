package persistence

import (
	"context"
	"errors"
	"fmt"
	"product-app/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

type IUserRepository interface {
	GetById(userId int64) (domain.User, error)
	GetByUsername(username string) (domain.User, error)
	GetByEmail(email string) (domain.User, error)
	AddUser(user domain.User) error
	UpdateUser(user domain.User) error
	DeleteById(userId int64) error
}

type UserRepository struct {
	dbPool *pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) IUserRepository {
	return &UserRepository{
		dbPool: dbPool,
	}
}

func (userRepository *UserRepository) GetById(userId int64) (domain.User, error) {
	ctx := context.Background()

	getByIdSql := `SELECT id, username, email, password, first_name, last_name, created_at, updated_at FROM users WHERE id = $1`
	queryRow := userRepository.dbPool.QueryRow(ctx, getByIdSql, userId)

	var user domain.User
	scanErr := queryRow.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("user not found with id %d: %w", userId, scanErr)
	}

	if scanErr != nil {
		return domain.User{}, fmt.Errorf("error while getting user with id %d: %w", userId, scanErr)
	}

	return user, nil
}

func (userRepository *UserRepository) GetByUsername(username string) (domain.User, error) {
	ctx := context.Background()

	getByUsernameSql := `SELECT id, username, email, password, first_name, last_name, created_at, updated_at FROM users WHERE username = $1`
	queryRow := userRepository.dbPool.QueryRow(ctx, getByUsernameSql, username)

	var user domain.User
	scanErr := queryRow.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("user not found with username %s: %w", username, scanErr)
	}

	if scanErr != nil {
		return domain.User{}, fmt.Errorf("error while getting user with username %s: %w", username, scanErr)
	}

	return user, nil
}

func (userRepository *UserRepository) GetByEmail(email string) (domain.User, error) {
	ctx := context.Background()

	getByEmailSql := `SELECT id, username, email, password, first_name, last_name, created_at, updated_at FROM users WHERE email = $1`
	queryRow := userRepository.dbPool.QueryRow(ctx, getByEmailSql, email)

	var user domain.User
	scanErr := queryRow.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("user not found with email %s: %w", email, scanErr)
	}

	if scanErr != nil {
		return domain.User{}, fmt.Errorf("error while getting user with email %s: %w", email, scanErr)
	}

	return user, nil
}

func (userRepository *UserRepository) AddUser(user domain.User) error {
	ctx := context.Background()

	insertUserSQL := `
		INSERT INTO users (username, email, password, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	var userId int64
	err := userRepository.dbPool.QueryRow(ctx, insertUserSQL,
		user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt).Scan(&userId)

	if err != nil {
		log.Printf("❌ Error inserting user: %v", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}

	log.Printf("✅ User inserted with ID: %d", userId)
	return nil
}

func (userRepository *UserRepository) UpdateUser(user domain.User) error {
	ctx := context.Background()

	updateSql := `UPDATE users SET username = $1, email = $2, first_name = $3, last_name = $4, updated_at = $5 WHERE id = $6`

	commandTag, err := userRepository.dbPool.Exec(ctx, updateSql, 
		user.Username, user.Email, user.FirstName, user.LastName, user.UpdatedAt, user.Id)

	if err != nil {
		return fmt.Errorf("error while updating user with id %d: %w", user.Id, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d not found", user.Id)
	}

	log.Printf("✅ User updated with id %d", user.Id)
	return nil
}

func (userRepository *UserRepository) DeleteById(userId int64) error {
	ctx := context.Background()

	deleteSql := `DELETE FROM users WHERE id = $1`

	commandTag, err := userRepository.dbPool.Exec(ctx, deleteSql, userId)

	if err != nil {
		log.Printf("ERROR: Error while deleting user with id %d: %v", userId, err)
		return fmt.Errorf("error while deleting user with id %d: %w", userId, err)
	}

	if commandTag.RowsAffected() == 0 {
		log.Printf("WARNING: User with id %d not found for deletion", userId)
		return fmt.Errorf("user with id %d not found", userId)
	}

	log.Printf("INFO: User deleted with id %d", userId)
	return nil
}