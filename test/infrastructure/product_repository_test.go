package infrastructure

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"os"
	"product-app/common/postgresql"
	"product-app/domain"
	"product-app/persistence"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

var productRepository persistence.IProductRepository
var dbPool *pgxpool.Pool
var ctx context.Context

func TestMain(m *testing.M) {
	ctx = context.Background()

	dbPool = postgresql.GetConnectionPool(ctx, postgresql.Config{
		Host:                  "localhost",
		Port:                  "6432",
		DbName:                "productapp_unit_test",
		UserName:              "postgres",
		Password:              "postgres",
		MaxConnections:        "10",
		MaxConnectionIdleTime: "30s",
	})

	productRepository = persistence.NewProductRepository(dbPool)
	fmt.Println("Before all tests")
	exitCode := m.Run()
	fmt.Println("After all tests")
	os.Exit(exitCode)
}

func setup(ctx context.Context, dbPool *pgxpool.Pool) {
	clear(ctx, dbPool)
	TestDataInitialize(ctx, dbPool)
}

func clear(ctx context.Context, dbPool *pgxpool.Pool) {
	_, err := dbPool.Exec(ctx, "TRUNCATE product_images, products RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Printf("Error truncating tables: %v", err)
	}
}

func TestGetAllProducts(t *testing.T) {
	setup(ctx, dbPool)

	expectedProducts := []domain.Product{
		{Id: 1, Name: "AirFryer", Price: 3000, Discount: 22, Store: "ABC TECH"},
		{Id: 2, Name: "Ütü", Price: 1500, Discount: 10, Store: "ABC TECH"},
		{Id: 3, Name: "Çamaşır Makinesi", Price: 10000, Discount: 15, Store: "ABC TECH"},
		{Id: 4, Name: "Lambader", Price: 2000, Discount: 0, Store: "Dekorasyon Sarayı"},
	}
	t.Run("GetAllProducts", func(t *testing.T) {
		actualProducts := productRepository.GettAllProducts()
		assert.Equal(t, 4, len(actualProducts))
		assert.Equal(t, expectedProducts, actualProducts)
	})

	clear(ctx, dbPool)
}

func TestGetAllProductsByStore(t *testing.T) {
	setup(ctx, dbPool)

	expectedProducts := []domain.Product{
		{Id: 1, Name: "AirFryer", Price: 3000, Discount: 22, Store: "ABC TECH"},
		{Id: 2, Name: "Ütü", Price: 1500, Discount: 10, Store: "ABC TECH"},
		{Id: 3, Name: "Çamaşır Makinesi", Price: 10000, Discount: 15, Store: "ABC TECH"},
	}

	t.Run("GetAllProductsByStore", func(t *testing.T) {
		actualProducts := productRepository.GetAllProductsByStore("ABC TECH")

		assert.Equal(t, len(expectedProducts), len(actualProducts), "Ürün sayısı eşleşmeli")
		assert.Equal(t, expectedProducts, actualProducts, "Ürünler eşleşmeli")
	})

	clear(ctx, dbPool)
}

func TestAddProduct(t *testing.T) {
	/*expectedProducts := []domain.Product{
		{
			Id:       1,
			Name:     "Kupa",
			Price:    100.0,
			Discount: 0.0,
			Store:    "Kırtasiye Merkezi",
		},
	}*/
	newProduct := domain.Product{
		Name:      "Kupa",
		Price:     100.0,
		Discount:  0.0,
		Store:     "Kırtasiye Merkezi",
		ImageUrls: []string{"https://example.com/iphone16-front.jpg"},
	}
	t.Run("AddProduct", func(t *testing.T) {
		productRepository.AddProduct(newProduct)
		actualProducts := productRepository.GettAllProducts()
		assert.Equal(t, 1, len(actualProducts))
	})

	clear(ctx, dbPool)
}

func TestGetProductById(t *testing.T) {
	setup(ctx, dbPool)
	t.Run("GetProductById", func(t *testing.T) {
		actualProduct, _ := productRepository.GetById(1)

		expectedProduct := domain.Product{
			Id:       1,
			Name:     "AirFryer",
			Price:    3000.0,
			Discount: 22.0,
			Store:    "ABC TECH",
		}

		assert.Equal(t, expectedProduct, actualProduct)

		_, err := productRepository.GetById(5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product not found with id 5")
	})
	clear(ctx, dbPool)
}

func TestDeleteById(t *testing.T) {
	setup(ctx, dbPool)
	t.Run("DeleteById", func(t *testing.T) {
		productRepository.DeleteById(1)
		_, err := productRepository.GetById(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product not found with id 1")
	})
	clear(ctx, dbPool)
}

func TestUpdatePrice(t *testing.T) {
	setup(ctx, dbPool)
	t.Run("UpdatePrice", func(t *testing.T) {
		productBeforeUpdate, _ := productRepository.GetById(1)
		assert.Equal(t, float32(3000.0), productBeforeUpdate.Price)
		productRepository.UpdatePrice(1, 4000.0)
		productAfterUpdate, _ := productRepository.GetById(1)
		assert.Equal(t, float32(4000.0), productAfterUpdate.Price)
	})
	clear(ctx, dbPool)
}

func TestDeleteAllProducts(t *testing.T) {
	setup(ctx, dbPool)

	t.Run("DeleteAllProducts", func(t *testing.T) {
		err := productRepository.DeleteAllProducts()
		assert.NoError(t, err)

		products := productRepository.GettAllProducts()
		assert.Len(t, products, 0, "Delete all Products")
	})

	clear(ctx, dbPool)
}
