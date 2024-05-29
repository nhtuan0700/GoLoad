package logic

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAccountParams struct {
	AccountName string
	Password    string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
}

type account struct {
	goquDatabase            *goqu.Database
	accountDataAccessor     database.AccountDataAccessor
	accountPasswordAccessor database.AccountPasswordDataAccessor
	logger                  *zap.Logger
	hashLogic               Hash
}

func NewAccount(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordAccessor database.AccountPasswordDataAccessor,
	logger *zap.Logger,
	hashLogic Hash,
) Account {
	return &account{
		goquDatabase:            goquDatabase,
		accountDataAccessor:     accountDataAccessor,
		accountPasswordAccessor: accountPasswordAccessor,
		logger:                  logger,
		hashLogic:               hashLogic,
	}
}

func (a *account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	accountNameTaken, err := a.isAccountAccountNameTaken(ctx, params.AccountName)

	if err != nil {
		return CreateAccountOutput{}, status.Error(codes.Internal, "failed to check if account name is taken")
	}

	if accountNameTaken {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name is already taken")
	}

	var accountID uint64
	txErr := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountID, err := a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})
		if err != nil {
			return err
		}

		hashedPassword, hasErr := a.hashLogic.Hash(ctx, params.Password)
		if hasErr != nil {
			return hasErr
		}

		err = a.accountPasswordAccessor.CreateAccountPassword(ctx, database.AccountPassword{
			OfAccountID: accountID,
			Hash: hashedPassword,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return CreateAccountOutput{}, txErr
	}
	return CreateAccountOutput{
		ID: accountID,
		AccountName: params.AccountName,
	}, nil
}

func (a *account) isAccountAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	_ = utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))

	// TODO with caching

	_, err := a.accountDataAccessor.GetAccountByAccountName(ctx, accountName)
	if err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			return false, nil
		}
	}

	return true, nil
}
