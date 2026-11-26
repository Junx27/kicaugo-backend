package event

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeRow struct{ err error }

func (r fakeRow) Scan(dest ...any) error { return r.err }

func (fakeRow) Close() {}

type fakeRows struct {
	nextCount int
	nextMax   int
	scanErr   error
	err       error
}

func (r *fakeRows) Close() {}

func (r *fakeRows) Conn() *pgx.Conn { return nil }

func (r *fakeRows) Err() error { return r.err }

func (r *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (r *fakeRows) Next() bool {
	if r.nextCount < r.nextMax {
		r.nextCount++
		return true
	}
	return false
}

func (r *fakeRows) Scan(dest ...any) error { return r.scanErr }

func (r *fakeRows) RawValues() [][]byte { return nil }

func (r *fakeRows) Values() ([]any, error) { return nil, nil }

func (r *fakeRows) TypeMap() *pgtype.Map { return nil }

type fakeQuerier struct {
	queryErr error
	rows     *fakeRows
	rowErr   error
}

func (f fakeQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f fakeQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return f.rows, nil
}

func (f fakeQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return fakeRow{err: f.rowErr}
}

func (f fakeQuerier) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}
