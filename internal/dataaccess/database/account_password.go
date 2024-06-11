package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TableNameAccountPassword = "account_passwords"

	ErrAccountPasswordNotFound = status.Error(codes.NotFound, "account password not found")
)

var (
	ColNameAccountPasswordOfAccountID = "of_account_id"
	ColNameAccountPasswordHash        = "hash"
)

type AccountPassword struct {
	OfAccountID uint64 `db:"of_account_id" goqu:"skipinsert,skipupdate"`
	Hash        string `db:"hash"`
}

type AccountPasswordDataAccessor interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	GetAccountPassword(ctx context.Context, ofAccountID uint64) (AccountPassword, error)
	UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	WithDatabase(database Database) AccountPasswordDataAccessor
}

type accountPasswordDataAccessor struct {
	database Database
	logger   *zap.Logger
}

func NewAccountPasswordDataAccessor(
	database *goqu.Database,
	logger *zap.Logger,
) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger:   logger,
	}
}

func (a *accountPasswordDataAccessor) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.Any("account_password", accountPassword))

	_, err := a.database.
		Insert(TableNameAccountPassword).
		Rows(goqu.Record{
			ColNameAccountPasswordOfAccountID: accountPassword.OfAccountID,
			ColNameAccountPasswordHash:        accountPassword.Hash,
		}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account password")
		return status.Error(codes.Internal, "failed to create account password")
	}

	return nil
}

func (a *accountPasswordDataAccessor) GetAccountPassword(ctx context.Context, ofAccountID uint64) (AccountPassword, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.Int("of_account_id", int(ofAccountID)))

	accountPassword := AccountPassword{}
	found, err := a.database.
		From(TableNameAccountPassword).
		Where(goqu.C(ColNameAccountPasswordOfAccountID).Eq(ofAccountID)).
		ScanStructContext(ctx, &accountPassword)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account password by of_account_id")
		return AccountPassword{}, status.Error(codes.Internal, "failed to get account by of_account_id")
	}

	if !found {
		logger.Warn("cannot find account password")
		return AccountPassword{}, ErrAccountPasswordNotFound
	}

	return accountPassword, nil
}

func (a *accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)

	_, err := a.database.
		Update(TableNameAccountPassword).
		Set(goqu.Record{
			ColNameAccountPasswordHash: accountPassword.Hash,
		}).
		Where(goqu.C(ColNameAccountPasswordOfAccountID).Eq(accountPassword.OfAccountID)).
		Executor().
		ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update account password")
		return status.Error(codes.Internal, "failed to update account password")
	}

	return nil
}

func (a *accountPasswordDataAccessor) WithDatabase(database Database) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger:   a.logger,
	}
}
