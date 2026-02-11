package repository

import (
	"context"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/infrastructure/database"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error)
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
}

type userRepository struct {
	db *database.Database
}

// NewUserRepository accepts *database.Database
func NewUserRepository(db *database.Database) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (email, password, created_at, updated_at) 
              VALUES ($1, $2, $3, $4) RETURNING id`

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Master for Create
	err := r.db.Master.QueryRow(ctx, query, user.Email, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, email, password, created_at, updated_at FROM users 
              WHERE email = $1 AND deleted_at IS NULL`

	var user entity.User
	// Slave for Read
	err := r.db.Slave.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error) {
	if order == "" {
		order = "created_at desc"
	}
	// Sanitize order to prevent SQL injection (basic check)
	// In a real app, use a whitelist of allowed columns

	offset := (page - 1) * limit

	// Count total
	var total int64
	countQuery := `SELECT count(*) FROM users WHERE deleted_at IS NULL`
	
	// Slave for Read
	if err := r.db.Slave.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// List users
	query := fmt.Sprintf(`SELECT id, email, created_at, updated_at FROM users 
                          WHERE deleted_at IS NULL 
                          ORDER BY %s LIMIT $1 OFFSET $2`, order)

	// Slave for Read
	rows, err := r.db.Slave.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	query := `SELECT id, email, password, created_at, updated_at FROM users 
              WHERE id = $1 AND deleted_at IS NULL`

	var user entity.User
	// Slave for Read
	err := r.db.Slave.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	query := `UPDATE users SET email = $1, password = $2, updated_at = $3 
              WHERE id = $4 AND deleted_at IS NULL`

	// Master for Update
	tag, err := r.db.Master.Exec(ctx, query, user.Email, user.Password, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found or deleted")
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	deletedAt := time.Now()
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	// Master for Delete
	tag, err := r.db.Master.Exec(ctx, query, deletedAt, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}
	return nil
}
