# Task Tracker Service

Simple REST API for managing tasks.

The service exposes `/tasks` and `/auth` endpoints.

## Routes

All task routes are under `/tasks`.

- **GET `/tasks/`**
  - Returns a paginated list of tasks.
  - Optional query parameters:
    - `page` (int, default `1`)
    - `limit` (int, default `10`)

- **GET `/tasks/{id}/`**
  - Returns a single task by its `id`.
  - Returns 404 if the task is not found.

- **POST `/tasks/`**
  - Creates a new task.
  - Returns the created task with generated `id`.
  - Example request body:
  
```json
{
  "name": "Task Name",
  "description": "Task Description"
}
```

- **PUT `/tasks/{id}/`**
  - Updates an existing task.
  - Expects JSON body like in POST.
  - Returns the updated task.

- **DELETE `/tasks/{id}/`**
  - Deletes a task by `id`.

## Auth API

Auth routes are under `/auth`.

- **POST `/auth/`**
  - Logs in a user and returns a JWT token.
  - Expects a JSON body:
    - `username` (string, required)
    - `password` (string, required)
  - Example request body:

```json
{
  "username": "user1",
  "password": "secret"
}
```

- **Authorization header**
  - Protected task routes (for example `GET /tasks/`) expect a bearer token.
  - Send the token from the login response in the `Authorization` header:

```http
Authorization: Bearer <token>
```

## Running the service

- Configure environment variables in `configs/.env`
- .env.example file is provided
- Run the API:

```bash
go run ./cmd/api
```
