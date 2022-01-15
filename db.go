package spansqlx

import (
	"context"
	"errors"
	"log"

	"cloud.google.com/go/spanner"
	"github.com/reiot101/spansqlx/internal"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	ErrBadConn = errors.New("scansqlx: bad connection")
	ErrNoRows  = errors.New("scansqlx: no rows")
)

type Options struct {
	database      string
	clientOptions []option.ClientOption
	clientConfig  *spanner.ClientConfig
}

type Option func(*Options) error

func WithDatabase(s string) Option {
	return func(o *Options) error {
		o.database = s
		return nil
	}
}

func WithClientOptions(opts ...option.ClientOption) Option {
	return func(o *Options) error {
		o.clientOptions = opts
		return nil
	}
}

func WithClientConfig(s *spanner.ClientConfig) Option {
	return func(o *Options) error {
		o.clientConfig = s
		return nil
	}
}

// DB is a wrapper around spanner.Client which keeps track of the options upon Open,
// used mostly to automatically bind named queries using the right bindvars.
type DB struct {
	opts Options
	db   *spanner.Client
}

// Open is the same as spanner.NewClient, but returns an *spansql.DB instead.
func Open(ctx context.Context, opts ...Option) (*DB, error) {
	// default options
	options := Options{
		database:      "projects/sandbox/instances/sandbox/databases/sandbox",
		clientOptions: []option.ClientOption{},
		clientConfig:  nil,
	}

	// apply options
	for i := range opts {
		if err := opts[i](&options); err != nil {
			return nil, err
		}
	}

	db := &DB{opts: options}
	return db, db.open(ctx)
}

// NewDb returns an DB instance.
func NewDb(ctx context.Context, db *spanner.Client) *DB {
	return &DB{
		db: db,
	}
}

// open spanner database connection
func (d *DB) open(ctx context.Context) error {
	var (
		db  *spanner.Client
		err error
	)

	if d.opts.clientConfig != nil {
		db, err = spanner.NewClientWithConfig(
			ctx,
			d.opts.database,
			*d.opts.clientConfig,
			d.opts.clientOptions...,
		)
	} else {
		db, err = spanner.NewClient(ctx, d.opts.database, d.opts.clientOptions...)
	}

	if err != nil {
		return err
	}

	d.db = db
	return d.Ping(ctx)
}

// Ping to a database and verify.
func (d *DB) Ping(ctx context.Context) error {
	var n int64
	if err := d.Get(ctx, &n, "SELECT 1"); err != nil {
		return err
	}
	if n == 0 {
		return ErrBadConn
	}
	return nil
}

// Select within a transaction.
// Any placeholder parameters are replaced with supplied args.
func (d *DB) Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	rows, err := d.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	return internal.ScanAll(rows, dest)
}

// SelectX within a transaction.
// Based spanner statement.
func (d *DB) SelectX(ctx context.Context, dest interface{}, stmt spanner.Statement) error {
	rows, err := d.QueryX(ctx, stmt)
	if err != nil {
		return err
	}
	return internal.ScanAll(rows, dest)
}

// Get within a transaction.
// Any placeholder parameters are replaced with supplied args.
// An error is returned if the result set is empty.
func (d *DB) Get(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	var row *spanner.Row

	stmt, err := internal.PrepareStmtAll(sql, args...)
	if err != nil {
		return err
	}

	err = forEach(ctx, d.db, func(iter *spanner.RowIterator) error {
		defer iter.Stop()

		if v, err := iter.Next(); err != nil && err != iterator.Done {
			return err
		} else {
			row = v
		}

		return nil
	}, stmt)

	if err != nil {
		return err
	}

	if row == nil {
		return ErrNoRows
	}

	return internal.ScanAny(row, dest)
}

// GetX within a transaction.
// Based spanner statement.
// An error is returned if the result set is empty.
func (d *DB) GetX(ctx context.Context, dest interface{}, stmt spanner.Statement) error {
	var row *spanner.Row

	err := forEach(ctx, d.db, func(iter *spanner.RowIterator) error {
		defer iter.Stop()

		if v, err := iter.Next(); err != nil && err != iterator.Done {
			return err
		} else {
			row = v
		}

		return nil
	}, stmt)

	if err != nil {
		return err
	}

	if row == nil {
		return ErrNoRows
	}

	return internal.ScanAny(row, dest)
}

// Query queries the database and returns an *spanner.Row slice.
// Any placeholder parameters are replaced with supplied args.
func (d *DB) Query(ctx context.Context, sql string, args ...interface{}) ([]*spanner.Row, error) {
	var rows []*spanner.Row

	stmt, err := internal.PrepareStmtAll(sql, args...)
	if err != nil {
		return nil, err
	}

	err = forEach(ctx, d.db, func(iter *spanner.RowIterator) error {
		return iter.Do(func(row *spanner.Row) error {
			rows = append(rows, row)
			return nil
		})
	}, stmt)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// QueryX queries the database and returns an *spanner.Row slice.
// Based spanner statement.
func (d *DB) QueryX(ctx context.Context, stmt spanner.Statement) ([]*spanner.Row, error) {
	var rows []*spanner.Row

	err := forEach(ctx, d.db, func(iter *spanner.RowIterator) error {
		return iter.Do(func(row *spanner.Row) error {
			rows = append(rows, row)
			return nil
		})
	}, stmt)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (d *DB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	stmt, err := internal.PrepareStmtAll(sql, args...)
	if err != nil {
		return err
	}

	// checks tx in context.
	if tx, ok := hasReadWriteTxContext(ctx); ok {
		return update(ctx, tx, stmt)
	}

	// exec the tx.
	if _, err := d.db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		return update(ctx, tx, stmt)
	}); err != nil {
		return nil
	}

	return nil
}

func (d *DB) ExecX(ctx context.Context, stmt spanner.Statement) error {
	// checks tx in context.
	if tx, ok := hasReadWriteTxContext(ctx); ok {
		return update(ctx, tx, stmt)
	}

	// exec the tx.
	if _, err := d.db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		return update(ctx, tx, stmt)
	}); err != nil {
		return nil
	}

	return nil
}

func (d *DB) NamedExec(ctx context.Context, sql string, arg interface{}) error {
	stmt, err := internal.PrepareStmtAny(sql, arg)
	if err != nil {
		return err
	}

	// checks tx in context.
	if tx, ok := hasReadWriteTxContext(ctx); ok {
		return update(ctx, tx, stmt)
	}

	// exec the tx.
	if _, err := d.db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		return update(ctx, tx, stmt)
	}); err != nil {
		return err
	}

	return nil
}

// Close the database connection
func (d *DB) Close() error {
	if d.db != nil {
		d.db.Close()
	}
	return nil
}

// TxPipeline is ReadWriteTransaction wrap.
func (d *DB) TxPipeline(ctx context.Context, callback func(ctx context.Context) error) error {
	_, err := d.db.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		return callback(SetTxContext(ctx, tx))
	})
	if err != nil {
		return err
	}
	return nil
}

// forEach within a transaction with row iterator
func forEach(ctx context.Context, db *spanner.Client, fn func(*spanner.RowIterator) error, stmt spanner.Statement) error {
	var it *spanner.RowIterator

	switch tx := hasTxContext(ctx).(type) {
	case *spanner.ReadOnlyTransaction:
		it = tx.Query(ctx, stmt)
	case *spanner.ReadWriteTransaction:
		it = tx.Query(ctx, stmt)
	default:
		it = db.Single().Query(ctx, stmt)
	}

	return fn(it)
}

// update within a transaction exec.
func update(ctx context.Context, tx *spanner.ReadWriteTransaction, stmt spanner.Statement) error {
	row, err := tx.Update(ctx, stmt)
	if err != nil {
		return err
	}
	log.Printf("update record(%d)s", row)
	return nil
}
