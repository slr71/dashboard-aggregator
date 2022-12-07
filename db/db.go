package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cyverse-de/dashboard-aggregator/config"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type GoquDatabase interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	From(cols ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ScanStruct(i interface{}, query string, args ...interface{}) (bool, error)
	ScanStructContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	ScanStructs(i interface{}, query string, args ...interface{}) error
	ScanStructsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	ScanVal(i interface{}, query string, args ...interface{}) (bool, error)
	ScanValContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error)
	ScanVals(i interface{}, query string, args ...interface{}) error
	ScanValsContext(ctx context.Context, i interface{}, query string, args ...interface{}) error
	Select(cols ...interface{}) *goqu.SelectDataset
	Trace(op, sqlString string, args ...interface{})
	Truncate(table ...interface{}) *goqu.TruncateDataset
	Update(table interface{}) *goqu.UpdateDataset
}

type Database struct {
	db     *sqlx.DB
	fullDB *goqu.Database
	goquDB GoquDatabase
}

func New(dbconn *sqlx.DB) *Database {
	goquDB := goqu.New("postgresql", dbconn)
	return &Database{
		db:     dbconn, // Used when a method needs direct access to sqlx for struct scanning.
		fullDB: goquDB, // Used when a method needs to use a method not defined in the GoquDatabase interface.
		goquDB: goquDB, // Used when a method needs to optionally support being run inside a transaction.
	}
}

type QuerySettings struct {
	hasLimit  bool
	limit     uint
	hasOffset bool
	offset    uint
	tx        *goqu.TxDatabase
}

// QueryOption defines the signature for functions that can modify a QuerySettings
// instance.
type QueryOption func(*QuerySettings)

// WithQueryLimit allows callers to add a limit SQL statement to a query.
func WithQueryLimit(limit uint) QueryOption {
	return func(s *QuerySettings) {
		s.hasLimit = true
		s.limit = limit
	}
}

// WithQueryOffset allows callers to add an offset SQL statement to a query.
func WithQueryOffset(offset uint) QueryOption {
	return func(s *QuerySettings) {
		s.hasOffset = true
		s.offset = offset
	}
}

// WithTX allows callers to use a query as part of a transaction.
func WithTX(tx *goqu.TxDatabase) QueryOption {
	return func(s *QuerySettings) {
		s.tx = tx
	}
}

func Connect(config *config.DatabaseConfiguration) (*sqlx.DB, error) {
	dbURI := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)
	dbconn := sqlx.MustConnect("postgres", dbURI)
	dbconn.SetMaxOpenConns(10)
	dbconn.SetConnMaxIdleTime(time.Minute)
	return dbconn, nil
}

func (d *Database) Healthz(ctx context.Context) error {
	db := d.goquDB
	v := goqu.T("version")
	query := db.From(v).
		Select(v.Col("version")).
		Order(v.Col("applied").
			Desc()).
		Limit(1).
		Executor()

	var result string
	found, err := query.ScanValContext(ctx, &result)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("no version found")
	}

	return nil
}
