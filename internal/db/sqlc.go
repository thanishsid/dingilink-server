package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PG interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type DBQ interface {
	Querier
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (DBT, error)
	Raw() *pgxpool.Pool
}

type DBT interface {
	Querier
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
	Raw() pgx.Tx
}

var _ DBQ = (*DB)(nil)

type DB struct {
	*Queries
	pg *pgxpool.Pool
}

var _ DBT = (*TxDB)(nil)

type TxDB struct {
	*Queries
	pgx.Tx
}

func (d *TxDB) Raw() pgx.Tx {
	return d.Tx
}

func (db DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (DBT, error) {
	tx, err := db.pg.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}

	return &TxDB{
		Queries: db.Queries.WithTx(tx),
		Tx:      tx,
	}, nil
}

func (d DB) Raw() *pgxpool.Pool {
	return d.pg
}

func NewDBQuerier(pg *pgxpool.Pool) *DB {
	return &DB{
		Queries: New(pg),
		pg:      pg,
	}
}

var PgTypesToRegister = []string{}

// Connect to postgres.
func Connect(ctx context.Context, dsn string, configModify func(conf *pgx.ConnConfig)) (*pgx.Conn, error) {

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	configModify(cfg)

	conn, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	for _, pt := range PgTypesToRegister {
		t, err := conn.LoadType(ctx, pt)
		if err != nil {
			return nil, err
		}
		conn.TypeMap().RegisterType(t)
	}

	return conn, nil
}

// Connect to postgres using a connection pool.
func ConnectPool(ctx context.Context, connstring string) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(connstring)
	if err != nil {
		return nil, err
	}

	conf.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, pt := range PgTypesToRegister {
			t, err := conn.LoadType(ctx, pt)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(t)
		}

		return nil
	}

	pg, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, err
	}

	return pg, nil
}
