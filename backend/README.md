# Backend (Go) - Authentication

This small backend provides simple JWT authentication using PostgreSQL (pgx).

Environment (see `.env.example`):

- `DATABASE_URL` - PostgreSQL DSN (e.g. `postgres://user:pass@localhost:5432/dbname?sslmode=disable`)
- `JWT_SECRET` - secret used to sign JWTs
- `ADDR` - server address (default `:8080`)

Quick run (requires Go 1.21+):

1. Create Postgres database and run the SQL in `migrations/001_create_users.sql`.
2. Copy `.env.example` to `.env` and set `DATABASE_URL` and `JWT_SECRET`.
3. From the `backend` folder:

```bash
go mod tidy
go run ./...
```

Endpoints

- POST /api/v1/auth/register  - body: { email, password, name }
- POST /api/v1/auth/signin     - body: { email, password } -> returns { token, user }
- GET  /api/v1/protected     - example protected endpoint (requires Authorization: Bearer <token>)
