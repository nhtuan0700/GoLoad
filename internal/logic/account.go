package logic

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/cache"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAccountParams struct {
	AccountName     string
	AccountPassword string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type AccountLogic interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
}

type accountLogic struct {
	goquDatabase                *goqu.Database
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	takenAccountNameCache       cache.TakeAccountName
	hashLogic                   HashLogic
	logger                      *zap.Logger
}

func NewAccountLogic(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	takenAccountNameCache cache.TakeAccountName,
	hashLogic HashLogic,
	logger *zap.Logger,
) AccountLogic {
	return &accountLogic{
		goquDatabase:                goquDatabase,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		takenAccountNameCache:       takenAccountNameCache,
		hashLogic:                   hashLogic,
		logger:                      logger,
	}
}

func (a accountLogic) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))

	isTakenAccountName, err := a.takenAccountNameCache.Has(ctx, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to get account name from cache, will fall back to database")
	} else if isTakenAccountName {
		return true, nil
	}

	_, err = a.accountDataAccessor.GetAccountByAccountName(ctx, accountName)
	if err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			return false, nil
		}
		return false, err
	}

	err = a.takenAccountNameCache.Add(ctx, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to set account name into taken set in cache")
	}

	return true, nil
}

func (a accountLogic) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	isAccountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
	if err != nil {
		return CreateAccountOutput{}, status.Error(codes.Internal, "failed to check if account name is taken")
	}

	if isAccountNameTaken {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name is already taken")
	}

	var accountID uint64
	txErr := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountID, err = a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			Name: params.AccountName,
		})
		if err != nil {
			return err
		}

		hashedPassword, err := a.hashLogic.Hash(ctx, params.AccountPassword)
		if err != nil {
			return err
		}

		err = a.accountPasswordDataAccessor.WithDatabase(td).CreateAccountPassword(ctx, database.AccountPassword{
			OfAccountID: accountID,
			Hash:        hashedPassword,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return CreateAccountOutput{}, err
	}

	return CreateAccountOutput{
		ID:          accountID,
		AccountName: params.AccountName,
	}, nil
}
