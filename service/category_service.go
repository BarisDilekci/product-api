package service

import (
	"errors"
	"product-app/domain"
	"product-app/persistence"
)

type ICategoryService interface {
	GetAllCategories() []domain.Category
	GetById(categoryId int64) (domain.Category, error)
	AddCategory(category domain.Category) error
	UpdateCategory(category domain.Category) error
	DeleteById(categoryId int64) error
}

type CategoryService struct {
	categoryRepository persistence.ICategoryRepository
}

func NewCategoryService(categoryRepository persistence.ICategoryRepository) ICategoryService {
	return &CategoryService{
		categoryRepository: categoryRepository,
	}
}

func (categoryService *CategoryService) GetAllCategories() []domain.Category {
	return categoryService.categoryRepository.GetAllCategories()
}

func (categoryService *CategoryService) GetById(categoryId int64) (domain.Category, error) {
	return categoryService.categoryRepository.GetById(categoryId)
}

func (categoryService *CategoryService) AddCategory(category domain.Category) error {
	if err := validateCategory(category); err != nil {
		return err
	}
	return categoryService.categoryRepository.AddCategory(category)
}

func (categoryService *CategoryService) UpdateCategory(category domain.Category) error {
	if err := validateCategory(category); err != nil {
		return err
	}
	return categoryService.categoryRepository.UpdateCategory(category)
}

func (categoryService *CategoryService) DeleteById(categoryId int64) error {
	return categoryService.categoryRepository.DeleteById(categoryId)
}

func validateCategory(category domain.Category) error {
	if err := validateNameWithRegex(category.Name, "category name is required"); err != nil {
		return err
	}

	if category.Description == "" {
		return errors.New("category description is required")
	}

	return nil
}