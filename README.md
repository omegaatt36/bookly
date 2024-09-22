# Bookly

Bookly is a monorepo, full-stack Go project that demonstrates a clean and simple backend architecture, coupled with a Y2K-inspired web application using Go templates and HTMX.

## Overview

This project showcases a modern approach to building web applications with Go, featuring:

- A robust backend API
- A lightweight frontend using Go templates and HTMX
- Clean architecture principles
- GORM for database operations
- JWT-based authentication

## Features

- User management (registration, login, profile updates)
- Account management
- Ledger entries for financial tracking
- RESTful API
- Server-side rendered web interface with HTMX for dynamic updates

## Project Structure

- `cmd/`: Contains the main entry points for different executables
- `app/`: Application layer, including API and web handlers
- `domain/`: Core business logic and interfaces
- `persistence/`: Database related code, including migrations and repositories
- `service/`: Service layer implementing business logic
- `deploy/`: Deployment configurations

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL or using docker
- Docker (optional, for local database setup)
- [Taskfile](https://taskfile.dev/)
- [Air](https://github.com/air-verse/air)

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/omegaatt36/bookly.git
   cd bookly
   ```

2. Set up the database:
   ```
   task setup-db
   ```

3. Run database migrations:
   ```
   task migrate-api
   ```

4. Start the API server:
   ```
   task run-api
   ```

5. In a separate terminal, start the web server:
   ```
   task run-web
   ```

6. Visit `http://localhost:8081` in your browser to access the web interface.

## Development

- Use `task fmt` to format the code
- Use `task lint` to run linters
- Use `task test` to run tests

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).
