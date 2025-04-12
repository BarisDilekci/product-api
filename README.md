# product-api

This project provides a simple product management API using the Go programming language and the Echo framework. Users can list products, retrieve product details by ID, add new products, update product prices, and delete products.

## Getting Started

Follow these steps to run the API in your local environment.

### Prerequisites

* **Go:** The Go programming language must be installed on your system. For installation instructions, visit [Go Downloads](https://go.dev/dl/).
* **Git:** Git must be installed to clone the repository. For installation instructions, visit [Git Downloads](https://git-scm.com/downloads).
* **curl** or **Postman:** You will need an HTTP client tool to test the API endpoints.

### Setup

1.  **Clone the Repository:**

    ```bash
    git clone <your_repository_github_address>
    cd <repository_name>
    ```

2.  **Install Dependencies:**

    ```bash
    go mod tidy
    ```

### Database Setup (For Testing Purposes)

This project uses a simple in-memory database for development and testing purposes. You do not need to set up the database manually. However, the `test/scripts/test_db.sh` script can be used to simulate a sample database setup for the test environment or for specific test scenarios.

**Warning:** This script is not suitable for production environments.

To run the script:

```bash
cd test/scripts
chmod +x test_db.sh
./test_db.sh
cd ../.. # Go back to the root directory
```

3.  **Run App:**
   
```bash
go run main.go
```

4.  **Curl:**
Get all products :
```bash
http://localhost:8080/api/v1/products
```
List Products by Store :
```bash
http://localhost:8080/api/v1/products?store=Store A
```
Get Product Details by ID :
```bash
http://localhost:8080/api/v1/products/1
```
Delete Product by ID :
```bash
http://localhost:8080/api/v1/products/2
```
Don't forget to change the method types
