package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"product-app/domain"
)

type IProductRepository interface {
	GettAllProducts() []domain.Product
	GetAllProductsByStore(storeName string) []domain.Product
	AddProduct(product domain.Product) error
	GetById(productId int64) (domain.Product, error)
	DeleteById(productId int64) error
	UpdatePrice(productId int64, newPrice float32) error
}

type ProductRepository struct {
	dbPool *pgxpool.Pool
}

func NewProductRepository(dbPool *pgxpool.Pool) IProductRepository {
	return &ProductRepository{
		dbPool: dbPool,
	}
}

func (productRepository *ProductRepository) GettAllProducts() []domain.Product {
	ctx := context.Background()
	productRows, err := productRepository.dbPool.Query(ctx, "SELECT * FROM products")

	if err != nil {
		log.Error("Error while getting all products %v", err)
		return []domain.Product{}
	}

	defer productRows.Close()

	products, err := extractProductFromRows(productRows)

	if err != nil {
		log.Errorf("Error while extracting products from rows: %v", err)
		return []domain.Product{}
	}
	return products

}
func (productRepository *ProductRepository) GetAllProductsByStore(storeName string) []domain.Product {
	ctx := context.Background()

	getProductByStoreNameSql := `Select * from products where store = $1`
	productRows, err := productRepository.dbPool.Query(ctx, getProductByStoreNameSql, storeName)

	if err != nil {
		log.Error("Error while getting all products %v", err)
		return []domain.Product{}
	}
	defer productRows.Close()
	products, err := extractProductFromRows(productRows)

	if err != nil {
		log.Errorf("Error while extracting products from rows: %v", err)
		return []domain.Product{}
	}
	return products
}
func (productRepository *ProductRepository) AddProduct(product domain.Product) error {
	ctx := context.Background()
	insert_sql := `Insert into products (name,price,discount,store) VALUES ($1,$2,$3,$4)`

	addNewProduct, err := productRepository.dbPool.Exec(ctx, insert_sql, product.Name, product.Price, product.Discount, product.Store)

	if err != nil {
		log.Printf("ERROR: Error while adding product '%s': %v", product.Name, err)
		return fmt.Errorf("failed to add product '%s': %w", product.Name, err)
	}

	if addNewProduct.RowsAffected() == 0 {
		log.Printf("WARNING: Product '%s' already exists", product.Name)
	}

	log.Info(fmt.Printf("Product added with %v", addNewProduct))
	return nil
}

func (productRepository *ProductRepository) GetById(productId int64) (domain.Product, error) {
	ctx := context.Background()

	getByIdSql := `Select * from products where id = $1`
	queryRow := productRepository.dbPool.QueryRow(ctx, getByIdSql, productId)

	var product domain.Product
	scanErr := queryRow.Scan(&product.Id, &product.Name, &product.Price, &product.Discount, &product.Store)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.Product{}, fmt.Errorf("product not found with id %d: %w", productId, scanErr)
	}

	if scanErr != nil {
		return domain.Product{}, fmt.Errorf("error while getting product with id %d: %w", productId, scanErr)
	}

	return product, nil
}
func (ProductRepository *ProductRepository) DeleteById(productId int64) error {
	ctx := context.Background()

	deleteSql := `Delete from products where id = $1`

	commandTag, err := ProductRepository.dbPool.Exec(ctx, deleteSql, productId)

	if err != nil {
		log.Printf("ERROR: Error while deleting product with id %d: %v", productId, err)
		return fmt.Errorf("error while deleting product with id %d: %w", productId, err)
	}

	if commandTag.RowsAffected() == 0 {
		log.Printf("WARNING: Product with id %d not found for deletion", productId)
		return fmt.Errorf("product with id %d not found", productId)
	}

	log.Printf("INFO: Product deleted with id %d", productId)
	return nil
}

func (productRepository *ProductRepository) UpdatePrice(productId int64, newPrice float32) error {
	ctx := context.Background()

	updateSql := `Update products set price = $1 where id = $2`

	_, err := productRepository.dbPool.Exec(ctx, updateSql, newPrice, productId)

	if err != nil {
		return errors.New(fmt.Sprintf("Error while updating product with id : %d", productId))
	}
	log.Info("Product %d price updated with new price %v", productId, newPrice)
	return nil
}
func extractProductFromRows(productRows pgx.Rows) ([]domain.Product, error) {
	var products = []domain.Product{}
	var id int64
	var name string
	var price float32
	var discount float32
	var store string

	for productRows.Next() {
		err := productRows.Scan(&id, &name, &price, &discount, &store)
		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, domain.Product{Id: id, Name: name, Price: price, Discount: discount, Store: store})
	}
	if err := productRows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return products, nil

}
