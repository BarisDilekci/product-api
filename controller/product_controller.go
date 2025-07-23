package controller

import (
	"net/http"
	"product-app/controller/request"
	"product-app/controller/response"
	"product-app/middleware"
	"product-app/service"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// ProductController handles HTTP requests for product operations
// It provides endpoints for CRUD operations on products with authentication support
type ProductController struct {
	productService service.IProductService
}

// NewProductController creates a new instance of ProductController
// Parameters:
//   - productService: Service interface for product business logic
//
// Returns:
//   - *ProductController: New controller instance
func NewProductController(productService service.IProductService) *ProductController {
	return &ProductController{productService: productService}
}

// RegisterRoutes registers all product-related HTTP routes
// Public routes (no authentication):
//   - GET /api/v1/products/:id - Get single product by ID
//   - GET /api/v1/products - Get all products (with optional store filter)
//
// Protected routes (JWT required):
//   - POST /api/v1/products - Create new product
//   - PUT /api/v1/products/:id - Update product price
//   - DELETE /api/v1/products/:id - Delete product by ID
//   - DELETE /api/v1/products/deleteAll - Delete all products
//   - GET /api/v1/products/my-products - Get current user's products
//
// Parameters:
//   - e: Echo instance for route registration
func (productController *ProductController) RegisterRoutes(e *echo.Echo) {
	// Public routes (no authentication required)
	e.GET("/api/v1/categories/:id/products", productController.GetProductsByCategoryId)
	e.GET("/api/v1/products/:id", productController.GetProductById)
	e.GET("/api/v1/products", productController.GetAllProducts)
	e.POST("/api/v1/products", productController.AddProduct)

	// Protected routes (authentication required)
	protected := e.Group("/api/v1/products", middleware.JWTMiddleware())
	protected.PUT("/:id", productController.UpdatePrice)
	protected.DELETE("/:id", productController.DeleteProductById)
	protected.DELETE("/deleteAll", productController.DeleteAllProducts)
}

func (productController *ProductController) GetProductsByCategoryId(c echo.Context) error {
	param := c.Param("id")
	categoryId, err := strconv.Atoi(param)

	if err != nil || categoryId <= 0 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: "Error: " + err.Error(),
		})
	}

	products, err := productController.productService.GetProductsByCategoryId(int64(categoryId))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			ErrorDescription: "Error: " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, response.ToResponseList(products))
}

func (productController *ProductController) GetProductById(c echo.Context) error {
	param := c.Param("id")
	productId, err := strconv.Atoi(param)

	if err != nil || productId <= 0 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: "Error: " + err.Error(),
		})
	}

	product, err := productController.productService.GetById(int64(productId))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			ErrorDescription: "Error:  " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, response.ToResponse(product))
}

func (productController *ProductController) GetAllProducts(c echo.Context) error {
	store := c.QueryParam("store")

	if len(store) == 0 {
		allProducts := productController.productService.GetAllProducts()
		return c.JSON(http.StatusOK, response.ToResponseList(allProducts))
	}
	productsWithGivenStore := productController.productService.GetAllProductsByStore(store)
	return c.JSON(http.StatusOK, response.ToResponseList(productsWithGivenStore))
}

func (productController *ProductController) AddProduct(c echo.Context) error {
	var addProductRequest request.AddProductRequest
	bindErr := c.Bind(&addProductRequest)
	if bindErr != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: bindErr.Error(),
		})
	}
	err := productController.productService.Add(addProductRequest.ToModel())

	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			ErrorDescription: err.Error(),
		})
	}
	return c.NoContent(http.StatusCreated)
}
func (productController *ProductController) UpdatePrice(c echo.Context) error {
	param := c.Param("id")
	productId, _ := strconv.Atoi(param)

	newPrice := c.QueryParam("newPrice")
	if len(newPrice) == 0 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: "Parameter newPrice is required!",
		})
	}
	convertedPrice, err := strconv.ParseFloat(newPrice, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: "NewPrice Format Disrupted!",
		})
	}
	productController.productService.UpdatePrice(int64(productId), float32(convertedPrice))
	return c.NoContent(http.StatusOK)
}

func (productController *ProductController) DeleteProductById(c echo.Context) error {
	param := c.Param("id")
	productId, _ := strconv.Atoi(param)
	err := productController.productService.DeleteById(int64(productId))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			ErrorDescription: err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}

func (productController *ProductController) DeleteAllProducts(c echo.Context) error {
	err := productController.productService.DeleteAllProducts()
	if err != nil {
		log.Printf("DeleteAllProducts error: %v", err)
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			ErrorDescription: err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}
