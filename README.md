# KicauGo Backend

Backend pilot untuk integrasi Flutter `rumahkicau.mobile`. Tahap ini hanya mencakup
autentikasi; kontrak OpenAPI dihasilkan Huma di `/openapi.json`.

## Konfigurasi

| Env | Wajib | Default | Keterangan |
|---|---|---|---|
| `DATABASE_URL` | ya | — | `postgres://user:pass@localhost:5432/kicaugo?sslmode=disable` |
| `JWT_SECRET` | ya | — | Minimal 32 karakter acak, jangan di-commit |
| `PORT` | tidak | `8080` | Port HTTP |

## Menjalankan

```bash
go run ./cmd/api
```

Migrasi, index, dan seeder dijalankan otomatis saat startup dari
`internal/database/migrations`, `internal/database/indexes`, dan
`internal/database/seeders`. Seeder bersifat idempotent sehingga aman dijalankan
berulang kali.

Akun demo: `test@gmail.com` / `test123@`.

## Struktur

```
cmd/api/                  # entrypoint + wiring config, database, service
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

## Flutter

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
