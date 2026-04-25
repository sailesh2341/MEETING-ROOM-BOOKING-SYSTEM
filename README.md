# Meeting Room Booking System

This project is a Go based meeting room booking system with REST APIs, PostgreSQL, JWT authentication, role based access, audit logs, and a small command line client.

## Features

- User registration and login with JWT tokens
- Admin and user roles
- Room create, list, update, and delete APIs
- Booking create, update, cancel, and history APIs
- Database level protection against overlapping room bookings
- Audit log entries for important user and admin actions
- CLI commands for common workflows
- Makefile targets for running, building, testing, and migrations
- Docker and GitHub Actions setup

## Project Structure

```text
cmd/launcher        API server entry point
cmd/booking-cli    CLI client entry point
config             environment based configuration
db                 database connection and migrations
db/models          shared response models
server             routing, handlers, and middleware
utils              JWT and password helpers
```

## Environment

Create a local environment file from `.env.example` or export the values directly.

```sh
export ADDRESS=:8080
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/meeting_room?sslmode=disable'
export JWT_SECRET='change_this_secret'
export ADMIN_CODE='admin123'
```

## Database

Create the database, then apply the schema.

```sh
createdb meeting_room
make migrate-up
```

To roll back the schema:

```sh
make migrate-down
```

## Run

```sh
make install
make run
```

The API runs on `http://localhost:8080` by default.

## Docker

Start PostgreSQL and the API:

```sh
docker compose up --build
```

Apply migrations from your machine:

```sh
make migrate-up
```

## CLI Examples

Register an admin:

```sh
go run ./cmd/booking-cli register -username admin -password secret123 -role admin -admin-code admin123
```

Login and copy the token from the response:

```sh
go run ./cmd/booking-cli login -username admin -password secret123
```

Create a room:

```sh
go run ./cmd/booking-cli create-room -token "$TOKEN" -name "Board Room" -capacity 10 -location "First Floor"
```

List available rooms:

```sh
go run ./cmd/booking-cli rooms -token "$TOKEN" -start "2026-05-01 10:00" -end "2026-05-01 11:00"
```

Book a room:

```sh
go run ./cmd/booking-cli book -token "$TOKEN" -room 1 -start "2026-05-01 10:00" -end "2026-05-01 11:00"
```

Cancel a booking:

```sh
go run ./cmd/booking-cli cancel -token "$TOKEN" -id 1
```

## API Endpoints

| Method | Endpoint | Access |
| --- | --- | --- |
| GET | `/health` | Public |
| POST | `/register` | Public |
| POST | `/login` | Public |
| GET | `/rooms` | User |
| GET | `/rooms/{id}` | User |
| POST | `/rooms` | Admin |
| PUT | `/rooms/{id}` | Admin |
| DELETE | `/rooms/{id}` | Admin |
| POST | `/bookings` | User |
| GET | `/bookings/me` | User |
| PUT | `/bookings/{id}` | Owner or admin |
| DELETE | `/bookings/{id}` | Owner or admin |
| GET | `/bookings` | Admin |
| GET | `/users` | Admin |
| PUT | `/users/{id}/role` | Admin |
| DELETE | `/users/{id}` | Admin |
| GET | `/audit-logs` | Admin |
