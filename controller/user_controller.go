package controller

import (
	"net/http"
	"product-app/middleware"
	"product-app/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService service.IUserService
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
}

func NewUserController(userService service.IUserService) *UserController {
	return &UserController{userService: userService}
}

func (userController *UserController) RegisterRoutes(e *echo.Echo) {
	// Public routes (no authentication required)
	e.POST("/api/v1/auth/register", userController.Register)
	e.POST("/api/v1/auth/login", userController.Login)
	
	// Protected routes (authentication required)
	protected := e.Group("/api/v1/users", middleware.JWTMiddleware())
	protected.GET("/:id", userController.GetUserById)
	protected.PUT("/:id", userController.UpdateUser)
	protected.DELETE("/:id", userController.DeleteUser)
}

func (userController *UserController) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := userController.userService.Register(req.Username, req.Email, req.Password, req.FirstName, req.LastName); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

func (userController *UserController) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	user, err := userController.userService.Login(req.UsernameOrEmail, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.Id, user.Username, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	// Return user info without password and include token
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   token,
		"user": map[string]interface{}{
			"id":         user.Id,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (userController *UserController) GetUserById(c echo.Context) error {
	param := c.Param("id")
	userId, err := strconv.Atoi(param)

	if err != nil || userId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	user, err := userController.userService.GetById(int64(userId))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	// Return user info without password
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         user.Id,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (userController *UserController) UpdateUser(c echo.Context) error {
	param := c.Param("id")
	userId, err := strconv.Atoi(param)

	if err != nil || userId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	var updateReq struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.Bind(&updateReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Get existing user
	user, err := userController.userService.GetById(int64(userId))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	// Update only the fields provided
	user.Username = updateReq.Username
	user.Email = updateReq.Email
	user.FirstName = updateReq.FirstName
	user.LastName = updateReq.LastName

	if err := userController.userService.UpdateUser(user); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User updated successfully",
	})
}

func (userController *UserController) DeleteUser(c echo.Context) error {
	param := c.Param("id")
	userId, err := strconv.Atoi(param)

	if err != nil || userId <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	if err := userController.userService.DeleteById(int64(userId)); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}