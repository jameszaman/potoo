package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type UserRepo struct {
	q *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: db.New(pool)}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash, firstName, lastName, phone string) (*db.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           newID(),
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &row, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &row, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*db.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &row, nil
}
