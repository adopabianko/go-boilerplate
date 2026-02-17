package repository

import (
	"context"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/infrastructure/database"
	"go-boilerplate/pkg/tracer"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, page, limit int, order string, timezone string) ([]entity.User, int64, error)
	GetByID(ctx context.Context, id string, timezone string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *database.Database
}

// NewUserRepository accepts *database.Database
func NewUserRepository(db *database.Database) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	ctx, span := tracer.StartSpan(ctx, "UserRepository.Create", "repository")
	defer span.End()

	query := `INSERT INTO users (email, password)
              VALUES ($1, $2) RETURNING id`

	// Master for Create
	err := r.db.Master.QueryRow(ctx, query, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := tracer.StartSpan(ctx, "UserRepository.GetByEmail", "repository")
	defer span.End()

	query := `SELECT id, email, password FROM users 
              WHERE email = $1 AND deleted_at IS NULL`

	var user entity.User
	// Slave for Read
	err := r.db.Slave.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, page, limit int, order string, timezone string) ([]entity.User, int64, error) {
	ctx, span := tracer.StartSpan(ctx, "UserRepository.List", "repository")
	defer span.End()

	if timezone == "" {
		timezone = "UTC"
	}

	if order == "" {
		order = "created_at desc"
	}

	offset := (page - 1) * limit

	// Count total
	var total int64
	countQuery := `SELECT count(*) FROM users WHERE deleted_at IS NULL`

	// Slave for Read
	if err := r.db.Slave.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// List users
	query := fmt.Sprintf(`SELECT id, email, created_at AT TIME ZONE '%s', updated_at AT TIME ZONE '%s' FROM users 
                          WHERE deleted_at IS NULL 
                          ORDER BY %s LIMIT $1 OFFSET $2`, timezone, timezone, order)

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

func (r *userRepository) GetByID(ctx context.Context, id string, timezone string) (*entity.User, error) {
	ctx, span := tracer.StartSpan(ctx, "UserRepository.GetByID", "repository")
	defer span.End()

	if timezone == "" {
		timezone = "UTC"
	}

	query := fmt.Sprintf(`SELECT id, email, password, created_at AT TIME ZONE '%s', updated_at AT TIME ZONE '%s' FROM users 
              WHERE id = $1 AND deleted_at IS NULL`, timezone, timezone)

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
	ctx, span := tracer.StartSpan(ctx, "UserRepository.Update", "repository")
	defer span.End()

	query := `UPDATE users SET email = $1, password = $2, updated_at = CURRENT_TIMESTAMP 
              WHERE id = $3 AND deleted_at IS NULL RETURNING id`

	// Master for Update
	err := r.db.Master.QueryRow(ctx, query, user.Email, user.Password, user.ID).Scan(&user.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found or deleted")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	ctx, span := tracer.StartSpan(ctx, "UserRepository.Delete", "repository")
	defer span.End()

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
