# Bookly

Bookly is a monorepo, full-stack Go project that demonstrates a clean and simple backend architecture, coupled with a Y2K-inspired web application using Go templates and HTMX.

## Overview

This project showcases a modern approach to building web applications with Go, featuring:

- A robust backend API
- A lightweight frontend using Go templates and HTMX
- Clean architecture principles
- SQLC/PGX for database operations
- JWT-based authentication
- Docker-based deployment

## Features

- User management (registration, login, profile updates)
- Account management
- Ledger entries for financial tracking
- RESTful API
- Server-side rendered web interface with HTMX for dynamic updates

## Project Structure

- `cmd/`: Contains the main entry points for different executables
  - `api/`: API service
  - `api-dbmigration/`: Database migration tool
  - `web/`: Web frontend service
- `app/`: Application layer, including API and web handlers
- `domain/`: Core business logic and interfaces
- `persistence/`: Database related code, including migrations and repositories
- `service/`: Service layer implementing business logic
- `deploy/`: Deployment configurations
- `Dockerfile.*`: Docker build files for different services

## Getting Started

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- [Taskfile](https://taskfile.dev/)

### Setup and Running

1. Clone the repository:
   ```
   git clone https://github.com/omegaatt36/bookly.git
   cd bookly
   ```

2. Start all services using Docker Compose:
   ```
   task dev
   ```

   This command will:
   - Set up the PostgreSQL database
   - Run database migrations
   - Start the API server
   - Start the web server
   - Start Adminer for database management

3. Access the services:
   - Web interface: `http://localhost:3000`
   - API: `http://localhost:8080`
   - Adminer (database management): `http://localhost:9527`

4. Register a new user via command line:
  ```shell
  curl -X POST 'http://localhost:8080/internal/auth/register' \
    -H 'Content-Type: application/json' \
    -H 'INTERNAL-TOKEN: secret' \
    -d '{"email":"tester","password":"tester"}'
  ```

## Development

- Use `task fmt` to format the code
- Use `task lint` to run linters
- Use `task test` to run tests
- Use `task live-api` to run the API with live reloading (powered by [Air](https://github.com/air-verse/air))
- Use `task live-web` to run the web server with live reloading (powered by [Air](https://github.com/air-verse/air))

## Database Management

- To set up the database: `task setup-db`
- To remove the database: `task remove-db`
- To run migrations: `task migrate-api`

## Technology Stack

- Backend: Go 1.23
- Database: PostgreSQL 16
- Data Access: pgx/v5, SQLC
- Authentication: JWT
- Web Framework: Go standard library http + templates
- Frontend Interaction: HTMX
- Deployment: Docker, Docker Compose
- Logging: slog, zap
- Code Quality: golangci-lint, revive

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).