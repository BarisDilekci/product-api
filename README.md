# product-api

This project provides a simple product management API built with Go and the Echo framework. It allows users to list products, retrieve product details by ID, add new products, update product prices, and delete products.

---

## Getting Started

### Prerequisites

- **Go:** Installed on your system.  
  [Go Installation Guide](https://go.dev/dl/)
- **Git:** Required to clone the repository.  
  [Git Installation Guide](https://git-scm.com/downloads)
- **HTTP Client:** Tools like `curl` or `Postman` for testing API endpoints.

### Setup

1. Clone the repository:

    ```bash
    git clone <repository_github_address>
    cd <repository_name>
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

3. **Setup and Run Database with Docker (Required):**

This project uses a Docker container to run the database for development and testing purposes. Make sure Docker is installed and running on your system.

To start the database container and prepare the environment, run the provided setup script:

```bash
cd test/scripts
chmod +x test_db.sh
./test_db.sh
cd ../..
```

4. Run the application:

    ```bash
    go run main.go
    ```

---

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

| Method | Endpoint                  | Description                     | Parameters                                                      | Successful Response                     |
|--------|---------------------------|---------------------------------|-----------------------------------------------------------------|----------------------------------------|
| GET    | `/products/:id`            | Get product details by ID       | Path: `id` (int64)                                              | 200 OK - JSON product object           |
| GET    | `/products`                | List all products                | Query: `store` (optional) - Filter products by store name       | 200 OK - JSON array of products        |
| POST   | `/products`                | Add a new product               | JSON body: `{ name, price, discount, store, imageUrls }`        | 201 Created - Added product object     |
| PUT    | `/products/:id`            | Update product price by ID      | JSON body: `{ newPrice }`                                       | 200 OK - Updated product object        |
| DELETE | `/products/:id`            | Delete product by ID            | Path: `id` (int64)                                              | 204 No Content                         |
| DELETE | `/products/deleteAll`      | Delete all products             | -                                                               | 204 No Content                         |

---

## Usage Examples


```bash
{
  "id": 1,
  "name": "Sample Product",
  "price": 99.99,
  "discount": 10,
  "store": "Store A",
  "imageUrls": ["https://example.com/image1.jpg"]
}

```

## Additional Notes

- The `price` field is a floating-point number.
- The `discount` value must be between 0 and 70 percent.
- The API returns appropriate HTTP status codes and descriptive error messages.
- Ensure to use the correct HTTP method for each endpoint (GET, POST, PUT, DELETE).
- The in-memory database is for development and testing only and resets on application restart.
- For production environments, you should integrate a persistent database (e.g., PostgreSQL, MySQL).
- The test database setup script (test_db.sh) simulates test scenarios and is not intended for production use.

