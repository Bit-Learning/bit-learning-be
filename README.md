# Bit Learning API

[![CI](https://github.com/lcaohoanq/bit-learning-be-v2/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/lcaohoanq/bit-learning-be-v2/actions/workflows/ci.yml)

Backend Go modular monolith cho authentication và user profile, dùng `net/http`,
chi, pgx, sqlc, goose, slog và OpenTelemetry. Luồng xử lý trong mỗi feature là
`Handler → Service → Repository → sqlc`; `cmd/api` chỉ khởi tạo dependency và
HTTP server.

## Chạy local

Yêu cầu Go 1.24+ và Docker.

```bash
cp .env.example .env
docker compose up -d postgres
make run
```

`make run` tự đọc và export các biến trong `.env`. Nếu chạy trực tiếp
`go run ./cmd/api`, Go không tự động đọc file `.env`; khi đó cần export biến
trong shell trước.

Mặc định ứng dụng tự chạy goose migrations (`AUTO_MIGRATE=true`). Nếu muốn chạy
thủ công, cài `goose`, đặt `AUTO_MIGRATE=false`, rồi dùng `make migrate-up`.

### Developer workflow

Cài [Air](https://github.com/air-verse/air), sau đó chạy API với hot reload:

```bash
make dev
```

Lệnh này tự chuẩn bị Bruno local environment, khởi động PostgreSQL, đọc `.env`,
rồi rebuild/restart API mỗi khi file Go hoặc OpenAPI thay đổi. API reference
tương tác bằng Scalar có tại <http://localhost:8080/docs>; OpenAPI source có tại
<http://localhost:8080/openapi.yaml>.

Để gọi API bằng Bruno, mở thư mục `bruno` trong app và chọn environment `local`.
Request Register hoặc Login sẽ tự lưu `access_token` để các request Get Me và
Update Me sử dụng. File environment thật được ignore để token local không bị
commit; `local.example.bru` là template dùng chung.

## API

```bash
# Đăng ký
curl -s http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123","display_name":"Bit Learner"}'

# Đăng nhập
curl -s http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123"}'

# Đọc profile
curl -s http://localhost:8080/v1/users/me -H "Authorization: Bearer $TOKEN"

# Cập nhật profile
curl -s -X PATCH http://localhost:8080/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"display_name":"New name","bio":"Learning Go","avatar_url":"https://example.com/avatar.png"}'
```

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| GET | `/healthz` | Không | Liveness |
| POST | `/v1/auth/register` | Không | Tạo user và trả access token |
| POST | `/v1/auth/login` | Không | Đăng nhập |
| GET | `/v1/users/me` | Bearer | Lấy profile |
| PATCH | `/v1/users/me` | Bearer | Cập nhật profile |

Response thành công luôn nằm trong `data`:

```json
{
  "data": {
    "id": "7bb66b40-21d8-4d03-a8a5-d204ba759a4c",
    "email": "user@example.com",
    "display_name": "Bit Learner",
    "bio": "Learning Go",
    "avatar_url": "https://example.com/avatar.png"
  }
}
```

Register và login trả thêm `access_token`, `token_type`, `expires_at` cùng user
công khai trong `data`. Lỗi có cùng một contract và có chi tiết field khi lỗi
validation:

```json
{
  "error": {
    "code": "validation_error",
    "message": "request validation failed",
    "details": [{"field": "email", "rule": "email"}]
  }
}
```

Đặt `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318` để gửi trace qua OTLP/HTTP.
Log request là structured JSON và chứa `request_id`, `trace_id`, status, latency.

Sau khi sửa file trong `internal/database/queries`, chạy `make generate` để sqlc
tạo lại code trong `internal/database/db`.
