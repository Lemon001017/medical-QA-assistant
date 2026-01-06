# Medical QA Assistant Backend

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Create MySQL database:
```sql
CREATE DATABASE medical_qa;
```

3. Copy `.env.example` to `.env` and configure:
```bash
cp .env.example .env
```

4. Update `.env` with your database credentials, JWT secret, and optional OpenAI settings.

5. Run the server (the process will auto-load `.env` if present):
```bash
go run cmd/server/main.go
```

## API Endpoints

### Public Endpoints

- `POST /api/v1/auth/register` - Register a new user
  ```json
  {
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }
  ```

- `POST /api/v1/auth/login` - Login
  ```json
  {
    "username": "testuser",
    "password": "password123"
  }
  ```

### Protected Endpoints

All protected endpoints require `Authorization: Bearer <token>` header.
