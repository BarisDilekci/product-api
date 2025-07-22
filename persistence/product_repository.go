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
	// GetAllProductsByUser(userId int64) []domain.Product // Kullanıcı ID'si kaldırıldığı için bu metot kaldırılmalı veya güncellenmeli
	AddProduct(product domain.Product) error
	GetById(productId int64) (domain.Product, error)
	DeleteById(productId int64) error
	UpdatePrice(productId int64, newPrice float32) error
	DeleteAllProducts() error
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
	// SELECT sorgusundan user_id kaldırıldı
	productRows, err := productRepository.dbPool.Query(ctx, "SELECT id, name, price, description, discount, store, category_id FROM products")

	if err != nil {
		log.Errorf("Error while getting all products: %v", err) // Log mesajı güncellendi
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
        SELECT id, name, price, description, discount, store, category_id
        FROM products
        WHERE store = $1
    `

	productRows, err := productRepository.dbPool.Query(ctx, getProductByStoreNameSql, storeName)
	if err != nil {
		log.Errorf("❌ Error while querying products by store: %v", err) // Log mesajı güncellendi
		return []domain.Product{}
	}
	defer productRows.Close()

	var products []domain.Product

	for productRows.Next() {
		var p domain.Product
		// Scan işleminden user_id kaldırıldı
		err := productRows.Scan(&p.Id, &p.Name, &p.Price, &p.Description, &p.Discount, &p.Store, &p.CategoryID)
		if err != nil {
			log.Errorf("❌ Error while scanning product for store: %v", err) // Log mesajı güncellendi
			continue
		}

		// Görselleri çekme kısmı aynı kalır
		imageRows, err := productRepository.dbPool.Query(ctx, `
            SELECT image_urls FROM product_images
            WHERE product_id = $1
            ORDER BY display_order
        `, p.Id)
		if err != nil {
			log.Errorf("❌ Error while querying images for store product: %v", err) // Log mesajı güncellendi
			continue
		}

		var imageUrls []string
		for imageRows.Next() {
			var url string
			if err := imageRows.Scan(&url); err != nil {
				log.Errorf("❌ Failed to scan image url for store product: %v", err) // Log mesajı güncellendi
				continue
			}
			imageUrls = append(imageUrls, url)
		}
		imageRows.Close()

		p.ImageUrls = imageUrls
		products = append(products, p)
	}

	// Eğer GetAllProductsByUser metodu tamamen kaldırılacaksa, IProductRepository arayüzünden de silinmeli.
	// Şu anki senaryoda kaldırılması en mantıklı olanıdır.
	return products
}

// GetAllProductsByUser metodunu tamamen kaldırdık, çünkü Product modelinde UserId yok ve bu metot ona bağımlı.
// Eğer bu işlevselliğe başka bir şekilde ihtiyacınız varsa, Users tablosu üzerinden farklı bir sorgu stratejisi düşünebilirsiniz.
// func (productRepository *ProductRepository) GetAllProductsByUser(userId int64) []domain.Product {
//     // ... Bu metot kaldırıldığı için içeriği de silinmelidir
// }

func (productRepository *ProductRepository) AddProduct(product domain.Product) error {
	ctx := context.Background()

	// INSERT sorgusundan user_id kaldırıldı
	insertProductSQL := `
        INSERT INTO products (name, price, description, discount, store, category_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id;
    `

	var productId int64
	// QueryRow parametrelerinden product.UserID kaldırıldı
	err := productRepository.dbPool.QueryRow(ctx, insertProductSQL,
		product.Name, product.Price, product.Description, product.Discount, product.Store, product.CategoryID).Scan(&productId)

	if err != nil {
		log.Errorf("❌ Error inserting product: %v", err) // Log mesajı güncellendi
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
			log.Errorf("❌ Error inserting image for product %d: %v", productId, err) // Log mesajı güncellendi
			return fmt.Errorf("failed to insert image: %w", err)
		}
	}

	log.Printf("✅ Product and images added successfully")
	return nil
}

func (productRepository *ProductRepository) GetById(productId int64) (domain.Product, error) {
	ctx := context.Background()

	// SELECT sorgusundan user_id kaldırıldı
	getByIdSql := `SELECT id, name, price, description, discount, store, category_id FROM products WHERE id = $1`
	queryRow := productRepository.dbPool.QueryRow(ctx, getByIdSql, productId)

	var product domain.Product
	// Scan işleminden product.UserID kaldırıldı
	scanErr := queryRow.Scan(&product.Id, &product.Name, &product.Price, &product.Description, &product.Discount, &product.Store, &product.CategoryID)

	if errors.Is(scanErr, pgx.ErrNoRows) {
		return domain.Product{}, fmt.Errorf("product not found with id %d: %w", productId, scanErr)
	}

	if scanErr != nil {
		return domain.Product{}, fmt.Errorf("error while getting product with id %d: %w", productId, scanErr)
	}

	// Görselleri çek
	imageRows, err := productRepository.dbPool.Query(ctx, `
        SELECT image_urls FROM product_images
        WHERE product_id = $1
        ORDER BY display_order
    `, productId)
	if err != nil {
		return domain.Product{}, fmt.Errorf("error querying images for product %d: %w", productId, err)
	}

	var imageUrls []string
	for imageRows.Next() {
		var url string
		if err := imageRows.Scan(&url); err != nil {
			imageRows.Close() // Hata durumunda row'ları kapat
			return domain.Product{}, fmt.Errorf("error scanning image url for product %d: %w", productId, err)
		}
		imageUrls = append(imageUrls, url)
	}
	imageRows.Close() // Tüm row'lar okunduktan sonra kapat

	product.ImageUrls = imageUrls
	return product, nil
}

func (productRepository *ProductRepository) DeleteById(productId int64) error { // Metot receiver adı düzeltildi (ProductRepository -> productRepository)
	ctx := context.Background()

	deleteSql := `DELETE FROM products WHERE id = $1` // SQL komutu küçük harfle ve WHERE boşluk düzeltildi

	commandTag, err := productRepository.dbPool.Exec(ctx, deleteSql, productId) // Receiver adı düzeltildi

	if err != nil {
		log.Errorf("❌ Error while deleting product with id %d: %v", productId, err) // Log mesajı güncellendi
		return fmt.Errorf("error while deleting product with id %d: %w", productId, err)
	}

	if commandTag.RowsAffected() == 0 {
		log.Warnf("⚠️ Product with id %d not found for deletion", productId) // Log seviyesi ve mesajı güncellendi
		return fmt.Errorf("product with id %d not found", productId)
	}

	log.Infof("✅ Product deleted with id %d", productId) // Log seviyesi ve mesajı güncellendi
	return nil
}

func (productRepository *ProductRepository) DeleteAllProducts() error { // Metot receiver adı düzeltildi
	ctx := context.Background()
	deleteAllProductsSql := `DELETE FROM products` // SQL komutu küçük harfle

	commandTag, err := productRepository.dbPool.Exec(ctx, deleteAllProductsSql) // Receiver adı düzeltildi

	if err != nil {
		log.Errorf("❌ Error while deleting all products: %v", err) // Log mesajı güncellendi
		return fmt.Errorf("error while deleting all products: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		log.Warn("⚠️ No products found for deletion") // Log seviyesi ve mesajı güncellendi
		return fmt.Errorf("products not found for deletion")
	}

	log.Infof("✅ All products deleted successfully (%d rows affected)", commandTag.RowsAffected()) // Log seviyesi ve mesajı güncellendi
	return nil
}

func (productRepository *ProductRepository) UpdatePrice(productId int64, newPrice float32) error {
	ctx := context.Background()

	updateSql := `UPDATE products SET price = $1 WHERE id = $2` // SQL komutu küçük harfle ve SET boşluk düzeltildi

	_, err := productRepository.dbPool.Exec(ctx, updateSql, newPrice, productId)

	if err != nil {
		log.Errorf("❌ Error while updating product price for id %d: %v", productId, err) // Log mesajı ve hata formatı güncellendi
		return fmt.Errorf("error while updating product price with id %d: %w", productId, err)
	}
	log.Infof("✅ Product %d price updated to %v", productId, newPrice) // Log seviyesi ve mesajı güncellendi
	return nil
}

func (productRepository *ProductRepository) extractProductFromRows(ctx context.Context, productRows pgx.Rows) ([]domain.Product, error) {
	var products []domain.Product

	for productRows.Next() {
		var p domain.Product
		// Scan işleminden p.UserID kaldırıldı
		err := productRows.Scan(&p.Id, &p.Name, &p.Price, &p.Description, &p.Discount, &p.Store, &p.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}

		// Görselleri çekme kısmı aynı kalır
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
				imageRows.Close() // Hata durumunda row'ları kapat
				return nil, fmt.Errorf("error scanning image url: %w", err)
			}
			imageUrls = append(imageUrls, url)
		}
		imageRows.Close() // Tüm row'lar okunduktan sonra kapat

		p.ImageUrls = imageUrls
		products = append(products, p)
	}

	if err := productRows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return products, nil
}
