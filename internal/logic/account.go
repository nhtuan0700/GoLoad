package logic

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/cache"
	"github.com/nhtuan0700/GoLoad/internal/dataaccess/database"
	"github.com/nhtuan0700/GoLoad/internal/generated/grpc/go_load"
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

type CreateSessionParams struct {
	AccountName     string
	AccountPassword string
}

type CreateSessionOutput struct {
	Token   string
	Account *go_load.Account
}

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error)
}

type account struct {
	goquDatabase                *goqu.Database
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	takenAccountNameCache       cache.TakeAccountName
	hashLogic                   Hash
	tokenLogic                  Token
	logger                      *zap.Logger
}

func NewAccount(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	takenAccountNameCache cache.TakeAccountName,
	hashLogic Hash,
	tokenLogic Token,
	logger *zap.Logger,
) Account {
	return &account{
		goquDatabase:                goquDatabase,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		takenAccountNameCache:       takenAccountNameCache,
		hashLogic:                   hashLogic,
		tokenLogic:                  tokenLogic,
		logger:                      logger,
	}
}

func (a *account) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
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

func (a *account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
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

func (a *account) databaseAccountToProtoAccount(account database.Account) *go_load.Account {
	return &go_load.Account{
		Id:          account.ID,
		AccountName: account.Name,
	}
}

func (a *account) CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error) {
	existingAccount, err := a.accountDataAccessor.GetAccountByAccountName(ctx, params.AccountName)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	existingAccountPassword, err := a.accountPasswordDataAccessor.GetAccountPassword(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	isHashEqual, err := a.hashLogic.IsHashEqual(ctx, params.AccountPassword, existingAccountPassword.Hash)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	if !isHashEqual {
		return CreateSessionOutput{}, status.Error(codes.Unauthenticated, "incorrect password")
	}

	token, _, err := a.tokenLogic.GetToken(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}

	return CreateSessionOutput{
		Token:   token,
		Account: a.databaseAccountToProtoAccount(existingAccount),
	}, nil
}
