package persistence

import (
	"context"
	"errors"
	"fmt"
	"product-app/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

type ICategoryRepository interface {
	GetAllCategories() []domain.Category
	GetById(categoryId int64) (domain.Category, error)
	AddCategory(category domain.Category) error
	UpdateCategory(category domain.Category) error
	DeleteById(categoryId int64) error
}

type CategoryRepository struct {
	dbPool *pgxpool.Pool
}

func NewCategoryRepository(dbPool *pgxpool.Pool) ICategoryRepository {
	return &CategoryRepository{
		dbPool: dbPool,
	}
}

func (categoryRepository *CategoryRepository) GetAllCategories() []domain.Category {
	ctx := context.Background()
	categoryRows, err := categoryRepository.dbPool.Query(ctx, "SELECT id, name, description FROM categories")

	if err != nil {
		log.Errorf("Error while getting all categories %v", err)
		return []domain.Category{}
	}

	defer categoryRows.Close()
	var categories []domain.Category

	for categoryRows.Next() {
		var c domain.Category
		err := categoryRows.Scan(&c.Id, &c.Name, &c.Description)
		if err != nil {
			log.Errorf("Error while scanning category: %v", err)
			continue
		}
		categories = append(categories, c)
	}

	return categories
}

func (categoryRepository *CategoryRepository) GetById(categoryId int64) (domain.Category, error) {
	ctx := context.Background()

	getByIdSql := `SELECT id, name, description FROM categories WHERE id = $1`
	queryRow := categoryRepository.dbPool.QueryRow(ctx, getByIdSql, categoryId)

	var category domain.Category
	scanErr := queryRow.Scan(&category.Id, &category.Name, &category.Description)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.Category{}, fmt.Errorf("category not found with id %d: %w", categoryId, scanErr)
	}

	if scanErr != nil {
		return domain.Category{}, fmt.Errorf("error while getting category with id %d: %w", categoryId, scanErr)
	}

	return category, nil
}

func (categoryRepository *CategoryRepository) AddCategory(category domain.Category) error {
	ctx := context.Background()

	insertCategorySQL := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id;
	`

	var categoryId int64
	err := categoryRepository.dbPool.QueryRow(ctx, insertCategorySQL,
		category.Name, category.Description).Scan(&categoryId)

	if err != nil {
		log.Printf("❌ Error inserting category: %v", err)
		return fmt.Errorf("failed to insert category: %w", err)
	}

	log.Printf("✅ Category inserted with ID: %d", categoryId)
	return nil
}

func (categoryRepository *CategoryRepository) UpdateCategory(category domain.Category) error {
	ctx := context.Background()

	updateSql := `UPDATE categories SET name = $1, description = $2 WHERE id = $3`

	commandTag, err := categoryRepository.dbPool.Exec(ctx, updateSql, category.Name, category.Description, category.Id)

	if err != nil {
		return fmt.Errorf("error while updating category with id %d: %w", category.Id, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("category with id %d not found", category.Id)
	}

	log.Printf("✅ Category updated with id %d", category.Id)
	return nil
}

func (categoryRepository *CategoryRepository) DeleteById(categoryId int64) error {
	ctx := context.Background()

	deleteSql := `DELETE FROM categories WHERE id = $1`

	commandTag, err := categoryRepository.dbPool.Exec(ctx, deleteSql, categoryId)

	if err != nil {
		log.Printf("ERROR: Error while deleting category with id %d: %v", categoryId, err)
		return fmt.Errorf("error while deleting category with id %d: %w", categoryId, err)
	}

	if commandTag.RowsAffected() == 0 {
		log.Printf("WARNING: Category with id %d not found for deletion", categoryId)
		return fmt.Errorf("category with id %d not found", categoryId)
	}

	log.Printf("INFO: Category deleted with id %d", categoryId)
	return nil
}