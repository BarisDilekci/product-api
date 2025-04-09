package controller

import (
	"github.com/labstack/echo/v4"
	"product-app/service"
)

type ProductController struct {
	productService service.IProductService
}

func NewProductController(productService service.IProductService) *ProductController {
	return &ProductController{productService: productService}
}

func (productController *ProductController) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/v1/products/:id", productController.GetProductById)
	e.GET("/api/v1/products", productController.GetAllProducts)
	e.POST("/api/v1/products", productController.AddProduct)
	e.PUT("/api/v1/products/:id", productController.UpdatePrice)
	e.DELETE("/api/v1/products/:id", productController.DeleteProductById)
}

func (productController *ProductController) GetProductById(c echo.Context) error {
	return nil
}

func (productController *ProductController) GetAllProducts(c echo.Context) error {
	return nil
}

func (productController *ProductController) AddProduct(c echo.Context) error {
	return nil
}

func (productController *ProductController) UpdatePrice(c echo.Context) error {
	return nil
}
func (productController *ProductController) DeleteProductById(c echo.Context) error {
	return nil
}
