package service

import (
	"github.com/stretchr/testify/assert"
	"os"
	"product-app/domain"
	"product-app/service"
	"product-app/service/model"
	"testing"
)

var productService service.IProductService

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

func Test_ShouldGetAllProducts(t *testing.T) {
	t.Run("ShouldGetAllProducts", func(t *testing.T) {
		initialProducts := []domain.Product{
			{Id: 1, Name: "AirFryer", Price: 1000.0, Store: "ABC TECH", UserID: 1, CategoryID: 1},
			{Id: 2, Name: "Ütü", Price: 4000.0, Store: "ABC TECH", UserID: 1, CategoryID: 1},
		}
		fakeRepo := NewFakeProductRepository(initialProducts)
		productService := service.NewProductService(fakeRepo)

		actualProducts := productService.GetAllProducts()
		assert.Equal(t, 2, len(actualProducts))
	})
}

func Test_WhenNoValidationErrorOccurred_ShouldAddProduct(t *testing.T) {
	t.Run("WhenNoValidationErrorOccurred_ShouldAddProduct", func(t *testing.T) {
		fakeRepo := NewFakeProductRepository([]domain.Product{})
		productService := service.NewProductService(fakeRepo)

		err := productService.Add(model.ProductCreate{
			Name:       "Ütü",
			Price:      2000.0,
			Discount:   50,
			Store:      "ABC TECH",
			CategoryID: 1,
		}, 1) // userId parameter added

		assert.NoError(t, err, "Add metodu hata döndürdü")

		actualProducts := productService.GetAllProducts()
		assert.Equal(t, 1, len(actualProducts))
	})
}

func Test_WhenDiscountIsHigherThan70_ShouldNotAddProduct(t *testing.T) {
	t.Run("WhenDiscountIsHigherThan70_ShouldNotAddProduct", func(t *testing.T) {

		fakeRepo := NewFakeProductRepository([]domain.Product{})
		productService := service.NewProductService(fakeRepo)

		err := productService.Add(model.ProductCreate{
			Name:       "Ütü",
			Price:      2000.0,
			Discount:   75,
			Store:      "ABC TECH",
			CategoryID: 1,
		}, 1) // userId parameter added

		actualProducts := productService.GetAllProducts()
		assert.Equal(t, 0, len(actualProducts))

		assert.Error(t, err)
		assert.Equal(t, "discount must be between 0 and 70 percent", err.Error())
	})
}

func Test_FakeProductRepository_GetById(t *testing.T) {
	initialProducts := []domain.Product{
		{Id: 1, Name: "Product A", Price: 10.0, Store: "Store X", UserID: 1, CategoryID: 1},
		{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 1, CategoryID: 1},
	}
	fakeRepo := NewFakeProductRepository(initialProducts)

	t.Run("Should return product by ID if found", func(t *testing.T) {
		product, err := fakeRepo.GetById(2)
		assert.NoError(t, err)
		assert.Equal(t, initialProducts[1], product)
	})

	t.Run("Should return error if product not found", func(t *testing.T) {
		product, err := fakeRepo.GetById(3)
		assert.Error(t, err)
		assert.Equal(t, "Product not found with id 3", err.Error())
		assert.Equal(t, domain.Product{}, product)
	})
}

func Test_FakeProductRepository_DeleteById(t *testing.T) {
	t.Run("Should delete product by ID if found", func(t *testing.T) {
		initialProducts := []domain.Product{
			{Id: 1, Name: "Product A", Price: 10.0, Store: "Store X", UserID: 1, CategoryID: 1},
			{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 1, CategoryID: 1},
			{Id: 3, Name: "Product C", Price: 30.0, Store: "Store X", UserID: 1, CategoryID: 1},
		}
		fakeRepo := NewFakeProductRepository(initialProducts)

		err := fakeRepo.DeleteById(2)
		assert.NoError(t, err)
		products := fakeRepo.GettAllProducts()
		assert.Len(t, products, 2)
		assert.NotContains(t, products, domain.Product{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 1, CategoryID: 1})
	})

	t.Run("Should return error if product not found", func(t *testing.T) {
		initialProducts := []domain.Product{
			{Id: 1, Name: "Product A", Price: 10.0, Store: "Store X", UserID: 1, CategoryID: 1},
			{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 1, CategoryID: 1},
			{Id: 3, Name: "Product C", Price: 30.0, Store: "Store X", UserID: 1, CategoryID: 1},
		}
		fakeRepo := NewFakeProductRepository(initialProducts)

		err := fakeRepo.DeleteById(4)
		assert.Error(t, err)
		assert.Equal(t, "Product not found with id 4", err.Error())
		products := fakeRepo.GettAllProducts()
		assert.Len(t, products, 3)
	})
}

func Test_FakeProductRepository_UpdatePrice(t *testing.T) {
	initialProducts := []domain.Product{
		{Id: 1, Name: "Product A", Price: 10.0, Store: "Store X", UserID: 1, CategoryID: 1},
		{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 1, CategoryID: 1},
	}
	fakeRepo := NewFakeProductRepository(initialProducts)

	t.Run("Should update price if product found", func(t *testing.T) {
		newPrice := float32(25.0)
		err := fakeRepo.UpdatePrice(2, newPrice)
		assert.NoError(t, err)
		product, err := fakeRepo.GetById(2)
		assert.NoError(t, err)
		assert.Equal(t, newPrice, product.Price)
	})

	t.Run("Should return error if product not found", func(t *testing.T) {
		newPrice := float32(30.0)
		err := fakeRepo.UpdatePrice(3, newPrice)
		assert.Error(t, err)
		assert.Equal(t, "Product not found with id 3", err.Error())
		product, err := fakeRepo.GetById(1)
		assert.NoError(t, err)
		assert.Equal(t, float32(10.0), product.Price)
	})
}

// Yeni test: GetAllProductsByUser fonksiyonu için
func Test_FakeProductRepository_GetAllProductsByUser(t *testing.T) {
	initialProducts := []domain.Product{
		{Id: 1, Name: "Product A", Price: 10.0, Store: "Store X", UserID: 1, CategoryID: 1},
		{Id: 2, Name: "Product B", Price: 20.0, Store: "Store Y", UserID: 2, CategoryID: 1},
		{Id: 3, Name: "Product C", Price: 30.0, Store: "Store X", UserID: 1, CategoryID: 2},
	}
	fakeRepo := NewFakeProductRepository(initialProducts)

	t.Run("Should return products for specific user", func(t *testing.T) {
		products := fakeRepo.GetAllProductsByUser(1)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(1), products[0].UserID)
		assert.Equal(t, int64(1), products[1].UserID)
	})

	t.Run("Should return empty slice for user with no products", func(t *testing.T) {
		products := fakeRepo.GetAllProductsByUser(999)
		assert.Len(t, products, 0)
	})
}
