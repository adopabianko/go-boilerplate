package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	return gormDB, mock, err
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := newMockDB()
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)
	user := &entity.User{
		Email:    "test@example.com",
		Password: "hashed_password",
	}

	// Gorm internally uses a transaction for Create unless disabled or specific config.
	// But we passed SkipDefaultTransaction: true.
	// Expect insert.
	// Note: Gorm might query "returning" for ID/CreatedAt.

	const sqlInsert = `INSERT INTO "users" ("email","password","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5) RETURNING "id"`

	mock.ExpectQuery(regexp.QuoteMeta(sqlInsert)).
		WithArgs(user.Email, user.Password, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := newMockDB()
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)
	email := "test@example.com"

	const sqlSelect = `SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`

	rows := sqlmock.NewRows([]string{"id", "email", "password", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, email, "hashed_password", time.Now(), time.Now(), nil)

	mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).
		WithArgs(email, 1). // Limit 1
		WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List(t *testing.T) {
	db, mock, err := newMockDB()
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)
	page, limit := 1, 10
	order := "created_at desc"

	// Count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE "users"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// List query
	rows := sqlmock.NewRows([]string{"id", "email", "created_at"}).
		AddRow(1, "u1@example.com", time.Now()).
		AddRow(2, "u2@example.com", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL ORDER BY created_at desc LIMIT $1`)).
		WithArgs(limit).
		WillReturnRows(rows)

	users, total, err := repo.List(context.Background(), page, limit, order)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
