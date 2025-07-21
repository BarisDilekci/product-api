package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"product-app/common/app"
	"product-app/common/postgresql"
	"product-app/controller"
	"product-app/persistence"
	"product-app/service"
)

func main() {
	ctx := context.Background()
	e := echo.New()

	configurationManager := app.NewConfigurationManager()
	dbPool := postgresql.GetConnectionPool(ctx, configurationManager.PostgreSqlConfig)

	// Product
	productRepository := persistence.NewProductRepository(dbPool)
	productService := service.NewProductService(productRepository)
	productController := controller.NewProductController(productService)

	// Category
	categoryRepository := persistence.NewCategoryRepository(dbPool)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryController := controller.NewCategoryController(categoryService)

	// User
	userRepository := persistence.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	// Register routes
	productController.RegisterRoutes(e)
	categoryController.RegisterRoutes(e)
	userController.RegisterRoutes(e)

	e.Start("localhost:8080")
}
