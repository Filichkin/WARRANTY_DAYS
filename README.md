# WARRANTY_DAYS Monorepo

Repository structure:

- `backend/` — Go backend service (API, migrations, business logic).
- `frontend/` — frontend app (planned).

## Backend quick start

Run from `backend/`:

```bash
cd backend
go run ./cmd/api
```

Or run from repo root:

```bash
go -C backend run ./cmd/api
```

For full backend documentation, see `backend/README.md`.
