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
	_ "github.com/go-sql-driver/mysql"               // Import MySQL driver
)

type Database interface {
	Delete(table interface{}) *goqu.DeleteDataset
	Dialect() string
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	From(tables ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Logger(logger goqu.Logger)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	// i: A pointer to a struct
	ScanStruct(i interface{}, query string, args ...interface{}) (bool, error)
	// i: A pointer to a struct
	ScanStructContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	// i: A pointer to a slice of structs
	ScanStructs(i interface{}, query string, args ...interface{}) error
	// i: A pointer to a slice of structs
	ScanStructsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	// i: A pointer to a primitive value
	ScanVal(i interface{}, query string, args ...interface{}) (bool, error)
	// i: A pointer to a primitive value
	ScanValContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	// i: A pointer to a slice of primitive values
	ScanVals(i interface{}, query string, args ...interface{}) error
	// i: A pointer to a slice of primitive values
	ScanValsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	Select(cols ...interface{}) *goqu.SelectDataset
	// Logs a given operation with the specified sql and arguments
	Trace(op string, sqlString string, args ...interface{})
	Truncate(table ...interface{}) *goqu.TruncateDataset
	Update(table interface{}) *goqu.UpdateDataset
}

func InitializeAndMigrateUpDB(databaseConfig configs.Database, logger *zap.Logger) (*sql.DB, func(), error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Database,
	)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Printf("error connecting to the database: %+v\n", err)
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	migrator := NewMigrator(db, logger)
	err = migrator.Up(context.Background())

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to execute database up migration")
	}

	return db, cleanup, nil
}

func InitializeGoquDB(db *sql.DB) *goqu.Database {
	return goqu.New("mysql", db)
}
