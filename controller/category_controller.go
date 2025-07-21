package controller

import (
	"net/http"
	"product-app/domain"
	"product-app/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type CategoryController struct {
	categoryService service.ICategoryService
}

func NewCategoryController(categoryService service.ICategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

func (categoryController *CategoryController) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/v1/categories", categoryController.GetAllCategories)
	e.GET("/api/v1/categories/:id", categoryController.GetCategoryById)
	e.POST("/api/v1/categories", categoryController.AddCategory)
	e.PUT("/api/v1/categories/:id", categoryController.UpdateCategory)
	e.DELETE("/api/v1/categories/:id", categoryController.DeleteCategoryById)
}

func (categoryController *CategoryController) GetAllCategories(c echo.Context) error {
	categories := categoryController.categoryService.GetAllCategories()
	return c.JSON(http.StatusOK, categories)
}

func (categoryController *CategoryController) GetCategoryById(c echo.Context) error {
	param := c.Param("id")
	categoryId, err := strconv.Atoi(param)

	if err != nil || categoryId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid category ID",
		})
	}

	category, err := categoryController.categoryService.GetById(int64(categoryId))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, category)
}

func (categoryController *CategoryController) AddCategory(c echo.Context) error {
	var category domain.Category
	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := categoryController.categoryService.AddCategory(category); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Category created successfully",
	})
}

func (categoryController *CategoryController) UpdateCategory(c echo.Context) error {
	param := c.Param("id")
	categoryId, err := strconv.Atoi(param)

	if err != nil || categoryId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid category ID",
		})
	}

	var category domain.Category
	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	category.Id = int64(categoryId)

	if err := categoryController.categoryService.UpdateCategory(category); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Category updated successfully",
	})
}

func (categoryController *CategoryController) DeleteCategoryById(c echo.Context) error {
	param := c.Param("id")
	categoryId, err := strconv.Atoi(param)

	if err != nil || categoryId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid category ID",
		})
	}

	if err := categoryController.categoryService.DeleteById(int64(categoryId)); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}