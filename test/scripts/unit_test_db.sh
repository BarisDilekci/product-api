#!/bin/bash

# PostgreSQL test konteynerini başlat
docker run --name postgres-test-db -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 6432:5432 -d postgres:latest

echo "🟡 PostgreSQL test container starting..."
sleep 3

# Test veritabanını oluştur
docker exec -it postgres-test-db psql -U postgres -d postgres -c "CREATE DATABASE productapp_unit_test"
sleep 3
echo "✅ Test database 'productapp_unit_test' created"

# Test tablolarını oluştur
docker exec -it postgres-test-db psql -U postgres -d productapp_unit_test -c "
CREATE TABLE IF NOT EXISTS products (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  price DOUBLE PRECISION NOT NULL,
  description VARCHAR(350) NOT NULL,
  discount DOUBLE PRECISION,
  store VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS product_images (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  image_urls TEXT NOT NULL,
  is_main_image BOOLEAN DEFAULT FALSE,
  display_order INT DEFAULT 0
);"

sleep 2
echo "✅ Test tables created successfully in 'productapp_unit_test'"
