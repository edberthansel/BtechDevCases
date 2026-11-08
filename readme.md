# Wallet App: JWT Auth + Transfer

Register, login (JWT), protected route, and wallet transfer, designed for users with very poor connections.

**Stack:** Go (Gin) + PostgreSQL 16 (backend), Next.js + TypeScript + Tailwind (frontend), Docker Compose.

## Run

```bash
cp .env.example .env     # Windows CMD: copy .env.example .env
# edit .env and set JWT_SECRET to any long random string
docker compose up --build
```

- App: http://localhost:3000
- API: http://localhost:8080 (health check: `/health`)
- Reset all data: `docker compose down -v`

Requires Docker with Compose v2.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | required | Secret used to sign tokens |
| `JWT_EXPIRY_MINUTES` | 15 | Token lifetime (sliding) |
| `INITIAL_BALANCE` | 1000000 | Starting balance (IDR) for each new user |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | btech, btech_secret, btech_wallet | Database credentials |
| `CORS_ORIGIN` | http://localhost:3000 | Allowed frontend origin |
| `NEXT_PUBLIC_API_URL` | http://localhost:8080 | API address used by the browser (build time) |

## Features and API

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/register` | no | `email`, `password`, `confirmPassword`, with validation |
| POST | `/login` | no | Returns a JWT with `user_id`, `email`, `exp` |
| GET | `/me` | yes | `Hello [email], welcome back` |
| GET | `/wallet` | yes | Current balance (IDR) |
| POST | `/transfer` | yes | `recipient` (email), `amount`, `notes`; header `Idempotency-Key` required |
| GET | `/transactions` | yes | Incoming and outgoing history, paginated |

Users are logged out after 15 minutes of inactivity.
`/transfer` returns 201 (created), 200 (replayed), 400, 404, 409 (key reused with different data), 422 (insufficient funds).

## Design decisions

- **Poor connections:** transfers use a client-generated idempotency key. Retries reuse the same key, so a request that succeeded but whose response was lost is never applied twice. The frontend uses a 10s timeout, retries with backoff on network errors and 5xx (not 4xx), and disables the form while a request runs.
- **Money safety:** integer amounts (no floats); one DB transaction per transfer with `SELECT ... FOR UPDATE` locking in a consistent order; DB constraints (`balance >= 0`, `amount > 0`, unique idempotency key per sender, no self-transfer).
- **Sessions:** sliding expiration. Each authenticated response returns a fresh token in `X-Refreshed-Token`; the frontend also logs out after 15 minutes without user activity.
- **Security:** bcrypt password hashing, identical login error for unknown email and wrong password, explicit HS256 validation, app refuses to start without `JWT_SECRET`.

## Limitations

- Stateless tokens: no server-side logout or revocation, and no maximum session length.
- Token is kept in `localStorage` for simplicity; an `HttpOnly` cookie would be safer against XSS.
- **No automated tests due to time.** Register, login, transfer, duplicate submit with the same idempotency key, insufficient funds, and offline retry were verified manually (Postman and UI).