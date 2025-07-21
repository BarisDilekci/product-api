package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"product-app/domain"
	"product-app/persistence"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

type IUserService interface {
	Register(username, email, password, firstName, lastName string) error
	Login(usernameOrEmail, password string) (domain.User, error)
	GetById(userId int64) (domain.User, error)
	UpdateUser(user domain.User) error
	DeleteById(userId int64) error
}

type UserService struct {
	userRepository persistence.IUserRepository
}

func NewUserService(userRepository persistence.IUserRepository) IUserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) Register(username, email, password, firstName, lastName string) error {
	if err := validateRegistration(username, email, password, firstName, lastName); err != nil {
		return err
	}

	// Check if username already exists
	if _, err := userService.userRepository.GetByUsername(username); err == nil {
		return errors.New("username already exists")
	}

	// Check if email already exists
	if _, err := userService.userRepository.GetByEmail(email); err == nil {
		return errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := domain.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return userService.userRepository.AddUser(user)
}

func (userService *UserService) Login(usernameOrEmail, password string) (domain.User, error) {
	if usernameOrEmail == "" || password == "" {
		return domain.User{}, errors.New("username/email and password are required")
	}

	var user domain.User
	var err error

	// Try to find user by email first, then username
	if strings.Contains(usernameOrEmail, "@") {
		user, err = userService.userRepository.GetByEmail(usernameOrEmail)
	} else {
		user, err = userService.userRepository.GetByUsername(usernameOrEmail)
	}

	if err != nil {
		return domain.User{}, errors.New("invalid credentials")
	}

	// Verify password
	if !verifyPassword(password, user.Password) {
		return domain.User{}, errors.New("invalid credentials")
	}

	return user, nil
}

func (userService *UserService) GetById(userId int64) (domain.User, error) {
	return userService.userRepository.GetById(userId)
}

func (userService *UserService) UpdateUser(user domain.User) error {
	if err := validateUserUpdate(user); err != nil {
		return err
	}
	user.UpdatedAt = time.Now()
	return userService.userRepository.UpdateUser(user)
}

func (userService *UserService) DeleteById(userId int64) error {
	return userService.userRepository.DeleteById(userId)
}

func validateRegistration(username, email, password, firstName, lastName string) error {
	if err := validateNameWithRegex(username, "username is required"); err != nil {
		return err
	}

	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if err := validateEmail(email); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	if err := validateNameWithRegex(firstName, "first name is required"); err != nil {
		return err
	}

	if err := validateNameWithRegex(lastName, "last name is required"); err != nil {
		return err
	}

	return nil
}

func validateUserUpdate(user domain.User) error {
	if err := validateNameWithRegex(user.Username, "username is required"); err != nil {
		return err
	}

	if err := validateEmail(user.Email); err != nil {
		return err
	}

	if err := validateNameWithRegex(user.FirstName, "first name is required"); err != nil {
		return err
	}

	if err := validateNameWithRegex(user.LastName, "last name is required"); err != nil {
		return err
	}

	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}

// Password hashing using Argon2
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 64*1024, 1, 4, b64Salt, b64Hash), nil
}

func verifyPassword(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return false
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	testHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, testHash) == 1
}