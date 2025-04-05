package infrastructure

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"product-app/common/postgresql"
	"testing"
)
import "product-app/persistence"

var productRepository persistence.IProductRepository
var dbPool *pgxpool.Pool
var ctx context.Context

func TestMain(m *testing.M) {
	ctx = context.Background()

	dbPool = postgresql.GetConnectionPool(ctx, postgresql.Config{
		Host:                  "localhost",
		Port:                  "6432",
		DbName:                "productapp",
		UserName:              "postgress",
		Password:              "yourpassword",
		MaxConnections:        "10",
		MaxConnectionIdleTime: "30s",
	})

	productRepository = persistence.NewProductRepository(dbPool)
	fmt.Println("Before all tests")
	exitCode := m.Run()
	fmt.Println("After all tests")
	os.Exit(exitCode)
}

func TestGetAllProducts(t *testing.T) {
	fmt.Println("TestGetAllProducts")
	fmt.Println(productRepository)
	fmt.Println(dbPool)
}
