# KicauGo Backend

Backend pilot untuk integrasi Flutter `rumahkicau.mobile`. Tahap ini hanya mencakup
autentikasi dan Beranda; kontraknya ada di `docs/openapi.yaml`.

## Konfigurasi

| Env | Wajib | Default | Keterangan |
|---|---|---|---|
| `DATABASE_URL` | ya | — | `postgres://user:pass@localhost:5432/kicaugo?sslmode=disable` |
| `JWT_SECRET` | ya | — | Minimal 32 karakter acak, jangan di-commit |
| `PORT` | tidak | `8080` | Port HTTP |

## Menjalankan

```bash
psql "$DATABASE_URL" -f migrations/0001_init.sql
psql "$DATABASE_URL" -f migrations/0002_seed.sql
go run ./cmd/api
```

Akun demo hasil seed: `test@gmail.com` / `test123@`.

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
