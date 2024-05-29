package database

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TableNameAccountPasswords = goqu.T("account_passwords")
)

const (
	ColNameAccountPasswordsOfAccountID = "of_account_id"
	ColNameAccountPasswordsHash        = "hash"
)

type AccountPassword struct {
	OfAccountID uint64 `db:"of_account_id"`
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
	logger := utils.LoggerWithContext(ctx, a.logger)

	_, err := a.database.
		Insert(TableNameAccountPasswords).
		Rows(goqu.Record{
			ColNameAccountPasswordsOfAccountID: accountPassword.OfAccountID,
			ColNameAccountPasswordsHash:        accountPassword.Hash,
		}).
		Executor().
		Exec()

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account password")
		return status.Error(codes.Internal, "failed to create account password")
	}

	return nil
}

func (a *accountPasswordDataAccessor) GetAccountPassword(ctx context.Context, ofAccountID uint64) (AccountPassword, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.Uint64("of_account_id", ofAccountID))

	accountPassword := AccountPassword{}
	found, err := a.database.
		From(TableNameAccountPasswords).
		Where(goqu.C(ColNameAccountPasswordsOfAccountID).Eq(ofAccountID)).
		ScanStructContext(ctx, &accountPassword)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account password by id")
		return AccountPassword{}, status.Error(codes.Internal, "failed to create account password by id")
	}

	if !found {
		logger.Warn("cannot find account password by id")
		return AccountPassword{}, sql.ErrNoRows
	}

	return accountPassword, nil
}

func (a *accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)

	_, err := a.database.
		Update(TableNameAccountPasswords).
		Set(goqu.Record{ColNameAccountPasswordsHash: accountPassword.Hash}).
		Where(goqu.C(ColNameAccountPasswordsHash).Eq(accountPassword.OfAccountID)).
		Executor().
		Exec()

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update account password")
		return status.Error(codes.Internal, "failed to update account password")
	}

	return nil
}

func (a *accountPasswordDataAccessor) WithDatabase(database Database) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger: a.logger,
	}
}
