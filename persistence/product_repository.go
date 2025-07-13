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
	products, err := productRepository.extractProductFromRows(ctx, productRows)

	if err != nil {
		log.Errorf("Error while extracting products from rows: %v", err)
		return []domain.Product{}
	}
	return products

}
func (productRepository *ProductRepository) GetAllProductsByStore(storeName string) []domain.Product {
	ctx := context.Background()

	getProductByStoreNameSql := `
		SELECT id, name, price, discount, store
		FROM products
		WHERE store = $1
	`

	productRows, err := productRepository.dbPool.Query(ctx, getProductByStoreNameSql, storeName)
	if err != nil {
		log.Errorf("❌ Error while querying products: %v", err)
		return []domain.Product{}
	}
	defer productRows.Close()

	var products []domain.Product

	for productRows.Next() {
		var p domain.Product
		err := productRows.Scan(&p.Id, &p.Name, &p.Price, &p.Discount, &p.Store)
		if err != nil {
			log.Errorf("❌ Error while scanning product: %v", err)
			continue
		}

		// Görselleri çek
		imageRows, err := productRepository.dbPool.Query(ctx, `
			SELECT image_urls FROM product_images
			WHERE product_id = $1
			ORDER BY display_order
		`, p.Id)
		if err != nil {
			log.Errorf("❌ Error while querying images: %v", err)
			continue
		}

		var imageUrls []string
		for imageRows.Next() {
			var url string
			if err := imageRows.Scan(&url); err != nil {
				log.Errorf("❌ Failed to scan image url: %v", err)
				continue
			}
			imageUrls = append(imageUrls, url)
		}
		imageRows.Close()

		p.ImageUrls = imageUrls
		products = append(products, p)
	}

	return products
}

func (productRepository *ProductRepository) AddProduct(product domain.Product) error {
	ctx := context.Background()

	insertProductSQL := `
		INSERT INTO products (name, price, discount, store)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	var productId int64
	err := productRepository.dbPool.QueryRow(ctx, insertProductSQL,
		product.Name, product.Price, product.Discount, product.Store).Scan(&productId)

	if err != nil {
		log.Printf("❌ Error inserting product: %v", err)
		return fmt.Errorf("failed to insert product: %w", err)
	}

	log.Printf("✅ Product inserted with ID: %d", productId)

	insertImageSQL := `
		INSERT INTO product_images (product_id, image_urls, is_main_image, display_order)
		VALUES ($1, $2, $3, $4);
	`

	for i, url := range product.ImageUrls {
		isMain := (i == 0)
		_, err := productRepository.dbPool.Exec(ctx, insertImageSQL, productId, url, isMain, i)
		if err != nil {
			log.Printf("❌ Error inserting image: %v", err)
			return fmt.Errorf("failed to insert image: %w", err)
		}
	}

	log.Printf("✅ Product and images added successfully")
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
func (productRepository *ProductRepository) extractProductFromRows(ctx context.Context, productRows pgx.Rows) ([]domain.Product, error) {
	var products []domain.Product

	for productRows.Next() {
		var p domain.Product
		err := productRows.Scan(&p.Id, &p.Name, &p.Price, &p.Discount, &p.Store)
		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}

		// Görselleri çek
		imageRows, err := productRepository.dbPool.Query(ctx, `
			SELECT image_urls FROM product_images
			WHERE product_id = $1
			ORDER BY display_order
		`, p.Id)
		if err != nil {
			return nil, fmt.Errorf("error querying images for product %d: %w", p.Id, err)
		}

		var imageUrls []string
		for imageRows.Next() {
			var url string
			if err := imageRows.Scan(&url); err != nil {
				imageRows.Close()
				return nil, fmt.Errorf("error scanning image url: %w", err)
			}
			imageUrls = append(imageUrls, url)
		}
		imageRows.Close()

		p.ImageUrls = imageUrls
		products = append(products, p)
	}

	if err := productRows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return products, nil
}
