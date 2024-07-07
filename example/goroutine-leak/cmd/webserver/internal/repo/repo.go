package repo

import (
	"context"

	"github.com/google/uuid"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (_ *UserRepo) Create(ctx context.Context, name, email string) (string, error) {
	println("UserRepo.Create")
	return uuid.New().String(), nil
}
