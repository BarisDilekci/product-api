#!/bin/bash

# PostgreSQL konteynerini başlat
docker run --name postgres-test -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 6432:5432 -d postgres:latest

echo "PostgreSQL starting..."
sleep 3

# Database oluştur
docker exec -it postgres-test psql -U postgres -d postgres -c "CREATE DATABASE productapp"
sleep 3
echo "Database productapp created"

# Tabloları oluştur
docker exec -it postgres-test psql -U postgres -d productapp -c "
CREATE TABLE IF NOT EXISTS products (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  price DOUBLE PRECISION NOT NULL,
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
echo "Tables created successfully"
