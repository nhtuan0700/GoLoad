package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/doug-martin/goqu/v9"
	"github.com/nhtuan0700/GoLoad/internal/configs"
	"go.uber.org/zap"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql" // Import MySQL goqu dialect
	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	Delete(table any) *goqu.DeleteDataset
	Dialect() string
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	From(from ...any) *goqu.SelectDataset
	Insert(table any) *goqu.InsertDataset
	Logger(logger goqu.Logger)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ScanStruct(i any, query string, args ...any) (bool, error)
	ScanStructContext(ctx context.Context, i any, query string, args ...any) (bool, error)
	ScanStructs(i any, query string, args ...any) error
	ScanStructsContext(ctx context.Context, i any, query string, args ...any) error
	ScanVal(i any, query string, args ...any) (bool, error)
	ScanValContext(ctx context.Context, i any, query string, args ...any) (bool, error)
	ScanVals(i any, query string, args ...any) error
	ScanValsContext(ctx context.Context, i any, query string, args ...any) error
	Select(cols ...any) *goqu.SelectDataset
	Trace(op string, sqlString string, args ...any)
	Truncate(table ...any) *goqu.TruncateDataset
	Update(table any) *goqu.UpdateDataset
}

func InitializeGoquDB(db *sql.DB) *goqu.Database {
	return goqu.New("mysql", db)
}

func InitializeAndMigrateUpDB(database configs.Database, logger *zap.Logger) (*sql.DB, func(), error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		database.Username,
		database.Password,
		database.Host,
		database.Port,
		database.Database,
	)

	db, err := sql.Open(string(configs.DatabaseTypeMySQL), connectionString)
	if err != nil {
		log.Printf("error connecting to the database: %+v", err)
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	migrator := NewMigrator(db, logger)
	err = migrator.Up(context.Background())

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to execute database migration")
	}

	return db, cleanup, nil
}
