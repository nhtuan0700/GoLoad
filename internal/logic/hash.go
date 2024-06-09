package logic

import (
	"context"

	"github.com/nhtuan0700/GoLoad/internal/configs"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HashLogic interface {
	Hash(ctx context.Context, data string) (string, error)
}

type hashLogic struct {
	authConfig configs.Auth
}

func NewHash(
	authConfig configs.Auth,
) HashLogic {
	return &hashLogic{
		authConfig: authConfig,
	}
}

func (h hashLogic) Hash(_ context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.authConfig.Hash.Cost)

	if err != nil {
		return "", status.Error(codes.Internal, "failed to hash data")
	}

	return string(hashed), nil
}
