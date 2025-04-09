package service

import (
	"errors"
	"product-app/domain"
	"product-app/persistence"
	"product-app/service/model"
)

type IProductService interface {
	Add(productCreate model.ProductCreate) error
	DeleteById(productId int64) error
	GetById(productId int64) (domain.Product, error)
	UpdatePrice(productId int64, newPrice float32) error
	GetAllProducts() []domain.Product
	GetAllProductsByStore(storeName string) []domain.Product
}

type ProductService struct {
	productRepository persistence.IProductRepository
}

func NewProductService(productRepository persistence.IProductRepository) IProductService {
	return &ProductService{
		productRepository: productRepository,
	}
}
func (productService *ProductService) Add(productCreate model.ProductCreate) error {
	validateError := validateProductCreate(productCreate)
	if validateError != nil {
		return validateError
	}
	return productService.productRepository.AddProduct(domain.Product{
		Name:     productCreate.Name,
		Price:    productCreate.Price,
		Discount: productCreate.Discount,
		Store:    productCreate.Store,
	})
}

func (productService *ProductService) DeleteById(productId int64) error {
	return productService.productRepository.DeleteById(productId)
}
func (productService *ProductService) GetById(productId int64) (domain.Product, error) {
	return productService.productRepository.GetById(productId)
}
func (productService *ProductService) UpdatePrice(productId int64, newPrice float32) error {
	return productService.productRepository.UpdatePrice(productId, newPrice)
}
func (productService *ProductService) GetAllProducts() []domain.Product {
	return productService.productRepository.GettAllProducts()
}
func (productService *ProductService) GetAllProductsByStore(storeName string) []domain.Product {
	return productService.productRepository.GetAllProductsByStore(storeName)
}

func validateProductCreate(productCreate model.ProductCreate) error {
	if productCreate.Name == "" {
		return errors.New("product name is required")
	}
	if productCreate.Price <= 0 {
		return errors.New("product price is required")
	}
	if productCreate.Store == "" {
		return errors.New("product store is required")
	}
	if productCreate.Discount < 0 || productCreate.Discount > 70 {
		return errors.New("Discount must be between 0 and 70 percent")
	}
	return nil
}
