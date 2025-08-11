## Product API (Go + Echo)

Go ve Echo ile geliştirilmiş; ürün, kategori ve kullanıcı yönetimi sağlayan katmanlı bir REST API.

---

### Table of Contents

- Purpose and Features
- Architecture and Directory Layout
- Setup and Run
- Database Setup (Docker)
- Environment Variables and Configuration
- API Endpoints and Sample Requests
- Validation Rules and Error Format
- Running Tests
- Technologies

---

### Purpose and Features

- Product CRUD: list, get by id, create, update price, delete, delete all
- Category CRUD: list, get by id, create, update, delete
- User registration and login, user endpoints protected with JWT
- Persistent data layer on PostgreSQL (pgxpool)
- Layered architecture: Controller → Service → Repository → DB
- Unit and integration tests (with a test database)

---

### Architecture and Directory Layout

High-level flow: HTTP (Echo) → Controller → Service (business rules) → Repository (SQL) → PostgreSQL.

Key directories:

- `controller/`: HTTP endpoints (product, category, user)
- `middleware/`: JWT generation and validation
- `service/`: business rules and validation
- `persistence/`: PostgreSQL queries (pgxpool)
- `domain/`: data models (Product, Category, User)
- `common/`: app and PostgreSQL configuration
- `test/`: integration and service tests, database scripts

Entry point: `main.go` (starts Echo, wires dependencies, port: `localhost:8080`).

---

### Setup and Run

Prerequisites:

- Go (go.mod: `go 1.24`) – Go 1.21+ recommended
- Docker (for the database)
- cURL/Postman (to test endpoints)

Steps:

```bash
git clone <repo_url>
cd product-api
go mod tidy

# Start the database (Docker) – see section below

go run main.go
# Server: http://localhost:8080
```

---

### Database Setup (Docker)

By default the app connects to PostgreSQL at `localhost:6432`. To start a dev/test database with Docker:

```bash
cd test/scripts
chmod +x test_db.sh
./test_db.sh
cd ../..
```

What the script does:

- Starts a `postgres-test` container from `postgres:latest` (host port 6432)
- Creates the `productapp` database
- Creates `products`, `product_images`, `categories`, `users` tables and sets up relationships
- Inserts sample categories

Cleanup (optional):

```bash
docker rm -f postgres-test || true
```

Note: For integration tests there is a separate script `test/scripts/unit_test_db.sh` that creates a `productapp_unit_test` database.

---

### Environment Variables and Configuration

- JWT secret: `JWT_SECRET` (optional; if not set, a weak development default is used)
- Database configuration: hard-coded in `common/app/configuration_manager.go`. Defaults:
  - Host: `localhost`, Port: `6432`, User: `postgres`, Password: `postgres`, DB: `productapp`
  - Update this file if you plan to use different DB credentials/ports.

Example run:

```bash
export JWT_SECRET="your-super-secret-jwt-key-min-32-chars"
go run main.go
```

---

### API Endpoints

Base URL: `http://localhost:8080/api/v1`

#### Products

- GET `/products`
  - List all products. Optional `store` query to filter by store: `/products?store=ABC%20TECH`
- GET `/products/:id`
  - Get product by id
- GET `/categories/:id/products`
  - Get products by category
- POST `/products`
  - Create a new product (public)
- PUT `/products/:id`
  - Update product price (requires JWT)
- DELETE `/products/:id`
  - Delete a product (requires JWT)
- DELETE `/products/deleteAll`
  - Delete all products (requires JWT)

Request body (POST /products):

```json
{
  "name": "AirFryer",
  "price": 3000,
  "description": "AirFryer açıklaması",
  "discount": 10,
  "store": "ABC TECH",
  "image_urls": ["https://example.com/img1.jpg"],
  "category_id": 1
}
```

Response (GET /products/:id):

```json
{
  "name": "AirFryer",
  "price": 3000,
  "description": "AirFryer açıklaması",
  "discount": 10,
  "store": "ABC TECH",
  "image_urls": ["https://example.com/img1.jpg"],
  "category_id": 1
}
```

Note: The Product GET response intentionally omits the `id` field due to the current response mapping.

#### Categories

- GET `/categories`
- GET `/categories/:id`
- POST `/categories`
- PUT `/categories/:id`
- DELETE `/categories/:id`

Request body (POST/PUT):

```json
{
  "name": "Electronics",
  "description": "Electronic devices and gadgets"
}
```

#### Authentication and Users

- POST `/auth/register`
  - User registration
- POST `/auth/login`
  - Login and obtain a JWT token
- GET `/users/:id` (requires JWT)
- PUT `/users/:id` (requires JWT)
- DELETE `/users/:id` (requires JWT)

Login response example:

```json
{
  "message": "Login successful",
  "token": "<JWT>",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

JWT usage example (protected endpoints):

```bash
curl -H "Authorization: Bearer <JWT>" http://localhost:8080/api/v1/users/1
```

---

### Validation Rules

#### Product

- `name`: required, alphanumeric plus spaces
- `price`: must be > 0
- `store`: required, alphanumeric plus spaces
- `discount`: must be between 0 and 70

#### Category

- `name`: required
- `description`: required

#### User

- Registration: `username` (min 3), `email` (valid format), `password` (min 6), `first_name`, `last_name`
- Login: `username_or_email` and `password` are required
- Passwords are hashed with Argon2 and compared in constant time

---

### Error Format

- Product endpoints: `{ "errorDescription": "..." }`
- Category and user endpoints: `{ "error": "..." }`

HTTP status codes are returned according to the scenario (400/401/404/422/500 etc.).

---

### Running Tests

Initialize the test database for integration tests:

```bash
cd test/scripts
chmod +x unit_test_db.sh
./unit_test_db.sh
cd ../..
```

Then run tests:

```bash
go test ./...
```

Notes:

- Integration tests use `localhost:6432` and the `productapp_unit_test` database
- Tests truncate and re-seed table data

---

### Technologies

- Go, Echo (`github.com/labstack/echo/v4`)
- PostgreSQL, pgx/pgxpool
- JWT (`github.com/golang-jwt/jwt/v5`)
- Argon2 (password hashing)
- Testing: `testing`, `github.com/stretchr/testify`

---

### Tips

- Server: `http://localhost:8080`
- Default DB connection is `localhost:6432` (Docker script maps this port)
- Provide a strong `JWT_SECRET` via environment variable
- Change DB settings in: `common/app/configuration_manager.go`

For a detailed authentication flow, see `AUTHENTICATION_GUIDE.md` (Turkish).

