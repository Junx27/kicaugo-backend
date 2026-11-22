package database

import (
	"context"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Junx27/kicaugo-backend/internal/config"
)

type fakeExecer struct {
	queries []string
	err     error
}

func (f *fakeExecer) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	if f.err != nil {
		return pgconn.CommandTag{}, f.err
	}
	f.queries = append(f.queries, sql)
	return pgconn.CommandTag{}, nil
}

func TestSQLFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"0002_second.sql":      &fstest.MapFile{Data: []byte("SELECT 2;")},
		"0001_first.sql":       &fstest.MapFile{Data: []byte("SELECT 1;")},
		"notes.txt":            &fstest.MapFile{Data: []byte("ignore")},
		"sub/0003_ignored.sql": &fstest.MapFile{Data: []byte("SELECT 3;")},
	}

	got, err := sqlFiles(fsys)
	if err != nil {
		t.Fatalf("sqlFiles() error = %v", err)
	}
	want := []string{"0001_first.sql", "0002_second.sql"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sqlFiles() = %v, ingin %v", got, want)
	}
}

func TestSQLFilesError(t *testing.T) {
	if _, err := sqlFiles(errorFS{}); err == nil {
		t.Error("sqlFiles() ingin error saat ReadDir gagal")
	}
}

func TestApplyFiles(t *testing.T) {
	db := &fakeExecer{}
	source := sqlSource{
		label: "test",
		files: fstest.MapFS{
			"0002_second.sql": &fstest.MapFile{Data: []byte("SELECT 2;")},
			"0001_first.sql":  &fstest.MapFile{Data: []byte("SELECT 1;")},
		},
	}

	if err := applyFiles(context.Background(), db, source); err != nil {
		t.Fatalf("applyFiles() error = %v", err)
	}
	want := []string{"SELECT 1;", "SELECT 2;"}
	if !reflect.DeepEqual(db.queries, want) {
		t.Errorf("queries = %v, ingin %v", db.queries, want)
	}
}

func TestApplyFilesReadError(t *testing.T) {
	source := sqlSource{label: "test", files: readFailFS{}}
	if err := applyFiles(context.Background(), &fakeExecer{}, source); err == nil {
		t.Error("applyFiles() ingin error saat ReadFile gagal")
	}
}

func TestApplyFilesReadDirError(t *testing.T) {
	source := sqlSource{label: "test", files: errorFS{}}
	if err := applyFiles(context.Background(), &fakeExecer{}, source); err == nil {
		t.Error("applyFiles() ingin error saat ReadDir gagal")
	}
}

func TestApplyFilesExecError(t *testing.T) {
	db := &fakeExecer{err: errors.New("db down")}
	source := sqlSource{
		label: "test",
		files: fstest.MapFS{
			"0001_first.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
		},
	}

	if err := applyFiles(context.Background(), db, source); err == nil {
		t.Error("applyFiles() ingin error saat Exec gagal")
	}
}

func TestSetup(t *testing.T) {
	db := &fakeExecer{}
	if err := Setup(context.Background(), db); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if len(db.queries) == 0 {
		t.Error("Setup() tidak mengeksekusi file SQL")
	}
}

func TestSetupError(t *testing.T) {
	db := &fakeExecer{err: errors.New("db down")}
	if err := Setup(context.Background(), db); err == nil {
		t.Error("Setup() ingin error saat Exec gagal")
	}
}

func TestConnectInvalidURL(t *testing.T) {
	cfg := &config.Config{DatabaseURL: "://invalid"}
	if _, err := Connect(context.Background(), cfg); err == nil {
		t.Error("Connect() ingin error untuk URL tidak valid")
	}
}

func TestConnectContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &config.Config{DatabaseURL: "postgres://user:pass@127.0.0.1:5432/kicaugo?sslmode=disable"}
	if _, err := Connect(ctx, cfg); err == nil {
		t.Error("Connect() ingin error saat context dibatalkan")
	}
}

type errorFS struct{}

func (errorFS) Open(string) (fs.File, error) {
	return nil, errors.New("read error")
}

type readFailFS struct{}

func (readFailFS) Open(string) (fs.File, error) {
	return nil, errors.New("open error")
}

func (readFailFS) ReadDir(string) ([]fs.DirEntry, error) {
	return []fs.DirEntry{fakeDirEntry{name: "0001_bad.sql"}}, nil
}

func (readFailFS) ReadFile(string) ([]byte, error) {
	return nil, errors.New("read error")
}

type fakeDirEntry struct {
	name string
}

func (e fakeDirEntry) Name() string               { return e.name }
func (e fakeDirEntry) IsDir() bool                { return false }
func (e fakeDirEntry) Type() fs.FileMode          { return 0 }
func (e fakeDirEntry) Info() (fs.FileInfo, error) { return nil, errors.New("not implemented") }
