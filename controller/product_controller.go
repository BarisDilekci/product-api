package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"product-app/controller/request"
	"product-app/controller/response"
	"product-app/middleware"
	"product-app/service"
	"strconv"
)

// ProductController handles HTTP requests for product operations
// It provides endpoints for CRUD operations on products with authentication support
type ProductController struct {
	productService service.IProductService
}

// NewProductController creates a new instance of ProductController
// Parameters:
//   - productService: Service interface for product business logic
// Returns:
//   - *ProductController: New controller instance
func NewProductController(productService service.IProductService) *ProductController {
	return &ProductController{productService: productService}
}

// RegisterRoutes registers all product-related HTTP routes
// Public routes (no authentication):
//   - GET /api/v1/products/:id - Get single product by ID
//   - GET /api/v1/products - Get all products (with optional store filter)
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
	e.GET("/api/v1/products/:id", productController.GetProductById)
	e.GET("/api/v1/products", productController.GetAllProducts)
	
	// Protected routes (authentication required)
	protected := e.Group("/api/v1/products", middleware.JWTMiddleware())
	protected.POST("", productController.AddProduct)
	protected.PUT("/:id", productController.UpdatePrice)
	protected.DELETE("/:id", productController.DeleteProductById)
	protected.DELETE("/deleteAll", productController.DeleteAllProducts)
	protected.GET("/my-products", productController.GetMyProducts)
}

// GetProductById retrieves a single product by its ID
// This is a public endpoint - no authentication required
// 
// URL Parameters:
//   - id: Product ID (integer, must be positive)
//
// Returns:
//   - 200 OK: Product data with all fields including images
//   - 400 Bad Request: Invalid or missing ID parameter
//   - 404 Not Found: Product with given ID doesn't exist
//
// Example: GET /api/v1/products/123
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

// GetAllProducts retrieves all products with optional filtering by store
// This is a public endpoint - no authentication required
//
// Query Parameters:
//   - store (optional): Filter products by store name
//
// Returns:
//   - 200 OK: Array of products
//     - If no store parameter: Returns all products from all stores
//     - If store parameter provided: Returns only products from that store
//
// Examples:
//   - GET /api/v1/products - Get all products
//   - GET /api/v1/products?store=TechStore - Get products from TechStore only
func (productController *ProductController) GetAllProducts(c echo.Context) error {
	store := c.QueryParam("store")

	if len(store) == 0 {
		allProducts := productController.productService.GetAllProducts()
		return c.JSON(http.StatusOK, response.ToResponseList(allProducts))
	}
	productsWithGivenStore := productController.productService.GetAllProductsByStore(store)
	return c.JSON(http.StatusOK, response.ToResponseList(productsWithGivenStore))
}

// AddProduct creates a new product for the authenticated user
// This is a protected endpoint - JWT authentication required
//
// Request Body (JSON):
//   - name: Product name (required, alphanumeric + spaces)
//   - price: Product price (required, must be > 0)
//   - description: Product description (optional)
//   - discount: Discount percentage (optional, 0-70)
//   - store: Store name (required, alphanumeric + spaces)
//   - image_urls: Array of image URLs (optional)
//   - category_id: Category ID (required, must exist)
//
// Returns:
//   - 201 Created: Product successfully created (no content)
//   - 400 Bad Request: Invalid request body format
//   - 401 Unauthorized: Missing or invalid JWT token
//   - 422 Unprocessable Entity: Validation errors (price, discount, name, etc.)
//
// Example: POST /api/v1/products
// Authorization: Bearer <jwt_token>
// Body: {"name":"iPhone","price":1000,"store":"TechStore","category_id":1}
func (productController *ProductController) AddProduct(c echo.Context) error {
	// Get authenticated user ID from JWT token
	userIdInterface := c.Get("user_id")
	userId, ok := userIdInterface.(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			ErrorDescription: "Invalid user authentication",
		})
	}

	var addProductRequest request.AddProductRequest
	bindErr := c.Bind(&addProductRequest)
	if bindErr != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			ErrorDescription: bindErr.Error(),
		})
	}
	
	err := productController.productService.Add(addProductRequest.ToModel(), userId)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			ErrorDescription: err.Error(),
		})
	}
	return c.NoContent(http.StatusCreated)
}

// UpdatePrice updates the price of an existing product
// This is a protected endpoint - JWT authentication required
//
// URL Parameters:
//   - id: Product ID (integer, must exist)
//
// Query Parameters:
//   - newPrice: New price value (required, must be valid float > 0)
//
// Returns:
//   - 200 OK: Price successfully updated (no content)
//   - 400 Bad Request: Missing newPrice parameter or invalid price format
//   - 401 Unauthorized: Missing or invalid JWT token
//   - 404 Not Found: Product with given ID doesn't exist
//
// Example: PUT /api/v1/products/123?newPrice=1500.50
// Authorization: Bearer <jwt_token>
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

// DeleteProductById deletes a specific product by its ID
// This is a protected endpoint - JWT authentication required
// Note: Users can only delete their own products (enforced at service level)
//
// URL Parameters:
//   - id: Product ID (integer, must exist)
//
// Returns:
//   - 200 OK: Product successfully deleted (no content)
//   - 401 Unauthorized: Missing or invalid JWT token
//   - 404 Not Found: Product with given ID doesn't exist
//   - 403 Forbidden: User doesn't have permission to delete this product
//
// Example: DELETE /api/v1/products/123
// Authorization: Bearer <jwt_token>
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

// DeleteAllProducts deletes all products from the system
// This is a protected endpoint - JWT authentication required
// WARNING: This is a dangerous operation that removes ALL products from ALL users
// Should be used with extreme caution, typically only by admin users
//
// Returns:
//   - 200 OK: All products successfully deleted (no content)
//   - 401 Unauthorized: Missing or invalid JWT token
//   - 404 Not Found: Error occurred during deletion process
//
// Example: DELETE /api/v1/products/deleteAll
// Authorization: Bearer <jwt_token>
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

// GetMyProducts retrieves all products belonging to the authenticated user
// This is a protected endpoint - JWT authentication required
// Only returns products that were created by the current user
//
// Returns:
//   - 200 OK: Array of user's products (may be empty if user has no products)
//   - 401 Unauthorized: Missing or invalid JWT token
//
// Response includes all product fields:
//   - id, name, price, description, discount, store
//   - image_urls, category_id, user_id
//
// Example: GET /api/v1/products/my-products
// Authorization: Bearer <jwt_token>
func (productController *ProductController) GetMyProducts(c echo.Context) error {
	// Get authenticated user ID from JWT token
	userIdInterface := c.Get("user_id")
	userId, ok := userIdInterface.(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			ErrorDescription: "Invalid user authentication",
		})
	}

	products := productController.productService.GetAllProductsByUser(userId)
	return c.JSON(http.StatusOK, response.ToResponseList(products))
}
