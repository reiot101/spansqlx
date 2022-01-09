package spansqlx

import (
	"context"

	"cloud.google.com/go/spanner"
)

type txContextKey int8

const (
	// ReadWrite transaction
	rwTxContextKey txContextKey = iota + 1
	// ReadOnly transaction
	roTxContextKey
)

func SetTxContext(ctx context.Context, arg interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	switch tx := arg.(type) {
	case *spanner.ReadOnlyTransaction:
		ctx = context.WithValue(ctx, roTxContextKey, tx)
	case *spanner.ReadWriteTransaction:
		ctx = context.WithValue(ctx, rwTxContextKey, tx)
	}

	return ctx
}

func hasReadWriteTxContext(ctx context.Context) (*spanner.ReadWriteTransaction, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(rwTxContextKey).(*spanner.ReadWriteTransaction)
	if !ok {
		return nil, false
	}
	if tx == nil {
		return nil, false
	}
	return tx, true
}

func hasReadOnlyTxContext(ctx context.Context) (*spanner.ReadOnlyTransaction, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(rwTxContextKey).(*spanner.ReadOnlyTransaction)
	if !ok {
		return nil, false
	}
	if tx == nil {
		return nil, false
	}
	return tx, true
}

func hasTxContext(ctx context.Context) interface{} {
	if v, ok := hasReadOnlyTxContext(ctx); ok {
		return v
	}

	if v, ok := hasReadWriteTxContext(ctx); ok {
		return v
	}
	return nil
}
