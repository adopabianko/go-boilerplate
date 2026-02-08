package repository_test

import (
	"context"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func newMockDB(t *testing.T) (pgxmock.PgxPoolIface, error) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		return nil, err
	}
	return mock, nil
}

func TestUserRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := repository.NewUserRepository(mock)
	user := &entity.User{
		Email:    "test@example.com",
		Password: "hashed_password",
	}

	const sqlInsert = `INSERT INTO users (email, password, created_at, updated_at) 
              VALUES ($1, $2, $3, $4) RETURNING id`

	mock.ExpectQuery(regexp.QuoteMeta(sqlInsert)).
		WithArgs(user.Email, user.Password, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uint(1)))

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := repository.NewUserRepository(mock)
	email := "test@example.com"

	const sqlSelect = `SELECT id, email, password, created_at, updated_at FROM users 
              WHERE email = $1 AND deleted_at IS NULL`

	rows := pgxmock.NewRows([]string{"id", "email", "password", "created_at", "updated_at"}).
		AddRow(uint(1), email, "hashed_password", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := repository.NewUserRepository(mock)
	page, limit := 1, 10
	order := "created_at desc"

	// Count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM users WHERE deleted_at IS NULL`)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// List query
	rows := pgxmock.NewRows([]string{"id", "email", "created_at", "updated_at"}).
		AddRow(uint(1), "u1@example.com", time.Now(), time.Now()).
		AddRow(uint(2), "u2@example.com", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, created_at, updated_at FROM users 
                          WHERE deleted_at IS NULL 
                          ORDER BY created_at desc LIMIT $1 OFFSET $2`)).
		WithArgs(limit, 0).
		WillReturnRows(rows)

	users, total, err := repo.List(context.Background(), page, limit, order)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := repository.NewUserRepository(mock)
	id := uint(999)

	const sqlSelect = `SELECT id, email, password, created_at, updated_at FROM users 
              WHERE id = $1 AND deleted_at IS NULL`

	mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetByID(context.Background(), id)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
