package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
type Execer interface {
	Querier
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
type Binder interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}

type DB struct {
	*sqlx.DB
}
type DBPool struct {
	*pgx.ConnPool
}

// just to count the total amount of the transactions
var transactions int = 0

func PostgresConnection(connString string) (*DB, error) {

	logrus.Debug("Connecting to PostgreSQL DB with:", connString)

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {

		return nil, err
	}

	// db.SetMaxOpenConns(1000) // The default is 0 (unlimited)
	// db.SetMaxIdleConns(10)   // defaultMaxIdleConns = 2
	// db.SetConnMaxLifetime(0) // 0, connections are reused forever.

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {

		return nil, err
	}

	logrus.Debug("Successfully connected to postgress")

	p := &DB{db}

	return p, nil
}

func PostgresConnectionPool(connString string) (*DBPool, error) {

	logrus.Debug("Connecting to PostgreSQL DB with:", connString)

	connConfig, err := pgx.ParseConnectionString(connString)

	connPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		AfterConnect:   nil,
		MaxConnections: 20,
		AcquireTimeout: 30 * time.Second,
	})
	if err != nil {

		return nil, errors.Wrap(err, "Call to pgx.NewConnPool failed")
	}
	p := &DBPool{connPool}

	return p, nil

	// nativeDB := stdlib.OpenDBFromPool(connPool)
	// if err != nil {
	// 	connPool.Close()
	// 	return nil, errors.Wrap(err, "Call to stdlib.OpenFromConnPool failed")
	// }
	// db, err := sqlx.NewDb(nativeDB, "pgx"), nil
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Call to pgx.NewConnPool failed")
	// }
	// p := &DBPool{db}

	// return p, nil
	// return

	// db, err := sqlx.Connect("pgx", connString)
	// if err != nil {

	// 	return nil, err
	// }

	// // Open doesn't open a connection. Validate DSN data:
	// err = db.Ping()
	// if err != nil {

	// 	return nil, err
	// }

	// logrus.Debug("Successfully connected to postgress")

	// p := &DBPool{db}

	// return p, nil
}

func (conn *DB) Transact(txFunc func(*sqlx.Tx) error) (err error) {

	transactions++

	// https://stackoverflow.com/questions/16184238/database-sql-tx-detecting-commit-or-rollback
	// https://stackoverflow.com/questions/51912841/golang-transactional-api-design

	tx, err := conn.Beginx()

	logrus.Debugf("Transaction opened [%d]", transactions)

	if err != nil {

		return
	}

	defer func() {

		if p := recover(); p != nil {

			logrus.Debugf("Transaction rollback on recover [%d]", transactions)
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {

			logrus.Debugf("Transaction rollback on error [%d]", transactions)
			tx.Rollback() // err is non-nil; don't change it
		} else {

			logrus.Debugf("Transaction commit [%d]", transactions)
			err = tx.Commit() // err is nil; if Commit returns error update err
		}

	}()

	err = txFunc(tx)

	return
}
