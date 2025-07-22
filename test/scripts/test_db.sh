#!/bin/bash

# PostgreSQL konteynerini başlat
echo "Starting PostgreSQL container..."
docker run --name postgres-test -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 6432:5432 -d postgres:latest

# PostgreSQL'in tamamen başlaması için bekleyin
echo "Waiting for PostgreSQL to start..."
sleep 5 # Daha uzun bir bekleme süresi, veritabanının tamamen ayağa kalkmasını sağlar

# Database oluştur
echo "Creating database 'productapp'..."
docker exec -it postgres-test psql -U postgres -d postgres -c "CREATE DATABASE productapp"
sleep 2 # Komutun bitmesini bekleyin
echo "Database 'productapp' created."

# Tabloları ve İlişkileri Oluştur
echo "Creating tables and setting up relationships..."
docker exec -it postgres-test psql -U postgres -d productapp -c "
-- Products table (mevcut yapınız)
CREATE TABLE IF NOT EXISTS products (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  price DOUBLE PRECISION NOT NULL,
  description VARCHAR(350) NOT NULL,
  discount DOUBLE PRECISION,
  store VARCHAR(255) NOT NULL
  -- category_id burada doğrudan tanımlanabilir veya ALTER TABLE ile eklenebilir
);

-- Product Images table (mevcut yapınız)
CREATE TABLE IF NOT EXISTS product_images (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_urls TEXT NOT NULL,
  is_main_image BOOLEAN DEFAULT FALSE,
  display_order INT DEFAULT 0
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT, -- description alanı NULL olabilir
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users table (bu tablo kalsın, diğer yerlerde kullanılıyor)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Update products table to include category_id
-- Sadece category_id'yi ekleyin, user_id'yi değil
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id BIGINT;

-- Add foreign key constraints
ALTER TABLE products ADD CONSTRAINT fk_products_category
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- user_id ile ilgili Foreign Key ve Index tanımlarını KALDIRIN:
-- ALTER TABLE products ADD CONSTRAINT fk_products_user
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- CREATE INDEX IF NOT EXISTS idx_products_user_id ON products(user_id);

-- Create other indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
"
sleep 2
echo "Tables and relationships created successfully."

# Insert some sample categories
echo "Inserting sample categories..."
docker exec -it postgres-test psql -U postgres -d productapp -c "
INSERT INTO categories (name, description) VALUES
('Electronics', 'Electronic devices and gadgets'),
('Clothing', 'Fashion and apparel items'),
('Books', 'Books and educational materials'),
('Home & Garden', 'Home improvement and gardening supplies')
ON CONFLICT (name) DO NOTHING; -- Sadece zaten yoksa ekle
"
sleep 1
echo "Sample categories inserted."

echo "Database setup complete!"
echo "You can connect to PostgreSQL using: psql -h localhost -p 6432 -U postgres -d productapp"