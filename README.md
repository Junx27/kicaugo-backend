# KicauGo Backend

Backend pilot untuk integrasi Flutter `rumahkicau.mobile`. Tahap ini hanya mencakup
autentikasi; kontrak OpenAPI dihasilkan Huma di `/openapi.json`.

## Status SonarCloud

[![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)
[![Security](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)
[![Reliability](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)
[![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=coverage)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)
[![Duplication](https://sonarcloud.io/api/project_badges/measure?project=Junx27_kicaugo-backend&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=Junx27_kicaugo-backend)

> **SonarCloud Quality Gate: Passed** — Security 0 issue · Duplication 0% · Reliability 0 bug · Maintainability 0 code smell

## Konfigurasi

| Env | Wajib | Default | Keterangan |
|---|---|---|---|
| `DATABASE_URL` | ya | — | `postgres://<user>:<password>@localhost:5432/<db>?sslmode=disable` |
| `JWT_SECRET` | ya | — | Minimal 32 karakter acak, jangan di-commit |
| `PORT` | tidak | `8080` | Port HTTP |
| `CARTO_BASEMAP_KEY` | tidak | — | Wajib untuk fitur basemap; bila kosong endpoint tile membalas 503. Jangan di-commit. |
| `CARTO_BASEMAP_URL` | tidak | template Voyager CARTO | Template tile dengan placeholder `{s}`, `{z}`, `{x}`, `{y}`, `{r}` |
| `CARTO_TILE_CACHE_TTL` | tidak | `168h` | TTL cache tile di memori |
| `CARTO_TILE_CACHE_MAX_ENTRIES` | tidak | `20000` | Jumlah maksimum tile yang di-cache |
| `CARTO_TILE_RATE_PER_MIN` | tidak | `300` | Batas request tile per user per menit |

## Menjalankan

```bash
go run ./cmd/api
```

Migrasi, index, dan seeder dijalankan otomatis saat startup dari
`internal/database/migrations`, `internal/database/indexes`, dan
`internal/database/seeders`. Seeder bersifat idempotent sehingga aman dijalankan
berulang kali.

Akun demo dibuat oleh seeder (lihat `internal/database/seeders`); jangan gunakan kredensial default di produksi.

## Struktur

```
cmd/api/                  # entrypoint + wiring config, database, service
cmd/scaffold/             # generator fitur CRUD (make scaffold)
cmd/server/               # konfigurasi HTTP server, router, middleware
internal/config/          # konfigurasi environment
internal/database/        # koneksi, migrations, indexes, seeders
internal/entities/        # model domain (User, AuthResult, BirdCategory)
internal/features/auth/   # fitur autentikasi
internal/features/birdcategory/ # fitur kategori burung (GET)
internal/middlewares/     # middleware HTTP (autentikasi JWT)
internal/utils/errors/    # error domain dan format error HTTP
internal/utils/http/      # helper registrasi route dan error HTTP
internal/utils/jwt/       # penerbitan dan validasi token JWT
internal/utils/request/   # tipe input request
internal/utils/response/  # format dan tipe response API
```

## Scaffold Generator

Gunakan `make scaffold` untuk membuat fitur CRUD baru secara otomatis (SQL → entities → features → utils → wiring).

```bash
# Fitur baru (full CRUD)
make scaffold NAME=product FIELDS="Name:string,Price:int,Active:bool"

# Fitur baru (hanya operasi tertentu)
make scaffold NAME=order FIELDS="Total:int,Status:string" OPS="list,create"

# Tambah operasi ke fitur yang sudah ada (contoh: delete)
make scaffold NAME=birdcategory OPS=delete
```

Tipe field yang didukung: `string`, `int`, `int64`, `bool`, `float64`, `time`.

Nilai `OPS`:
- Shorthand: `crud` (default), `read` (`list`+`get`), `write` (`create`+`update`+`delete`)
- Eksplisit: `create,list,get,update,delete` (comma-separated)

Yang dihasilkan:
- `internal/database/migrations/000N_<domain>_table.sql`
- `internal/entities/<domain>.go`
- `internal/features/<domain>/` (repository, service, handler, test)
- `internal/utils/request/<domain>.go` dan `internal/utils/response/<domain>.go`
- Wiring otomatis ke `cmd/api/main.go` dan `cmd/server/server.go`

Generator menjalankan `gofmt` dan `go build` otomatis; rollback jika gagal. Operasi augment (fitur existing) menambahkan method/route via marker `// scaffold:*`.

```bash
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

 saat di emulator Android gunakan `http://10.0.2.2:8080`, bukan `localhost`.

## Verifikasi

```bash
go build ./...
go vet ./...
go test -race ./internal/...
```
