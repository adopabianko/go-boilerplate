package repository_test

import (
	"context"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repository"
	"regexp"
	"testing"
	"time"

	"go-boilerplate/internal/infrastructure/database"

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

	db := &database.Database{
		Master: mock,
		Slave:  mock,
	}
	repo := repository.NewUserRepository(db)
	user := &entity.User{
		Email:    "test@example.com",
		Password: "hashed_password",
	}

	const sqlInsert = `INSERT INTO users (email, password)
              VALUES ($1, $2) RETURNING id`

	mock.ExpectQuery(regexp.QuoteMeta(sqlInsert)).
		WithArgs(user.Email, user.Password).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).
			AddRow("019c514b-a933-74f2-8d08-a496675c66cf"))

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, "019c514b-a933-74f2-8d08-a496675c66cf", user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &database.Database{
		Master: mock,
		Slave:  mock,
	}
	repo := repository.NewUserRepository(db)
	email := "test@example.com"

	const sqlSelect = `SELECT id, email, password FROM users 
              WHERE email = $1 AND deleted_at IS NULL`

	rows := pgxmock.NewRows([]string{"id", "email", "password"}).
		AddRow("019c514b-a933-74f2-8d08-a496675c66cf", email, "hashed_password")

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

	db := &database.Database{
		Master: mock,
		Slave:  mock,
	}
	repo := repository.NewUserRepository(db)
	page, limit := 1, 10
	order := "created_at desc"

	// Count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM users WHERE deleted_at IS NULL`)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// List query
	rows := pgxmock.NewRows([]string{"id", "email", "created_at", "updated_at"}).
		AddRow("019c514b-a933-74f2-8d08-a496675c66cf", "u1@example.com", time.Now(), time.Now()).
		AddRow("019c514b-a933-74f2-8d08-a496675c66d0", "u2@example.com", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC' FROM users 
                          WHERE deleted_at IS NULL 
                          ORDER BY created_at desc LIMIT $1 OFFSET $2`)).
		WithArgs(limit, 0).
		WillReturnRows(rows)

	users, total, err := repo.List(context.Background(), page, limit, order, "UTC")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &database.Database{
		Master: mock,
		Slave:  mock,
	}
	repo := repository.NewUserRepository(db)
	id := "019c514b-a933-74f2-8d08-a496675c66cf"

	const sqlSelect = `SELECT id, email, password, created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC' FROM users 
              WHERE id = $1 AND deleted_at IS NULL`

	mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetByID(context.Background(), id, "UTC")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}
