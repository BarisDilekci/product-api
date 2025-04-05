#!/bin/bash

# PostgreSQL konteynerini başlat
docker run --name postgress-test -e POSTGRES_USER=postgress -e POSTGRES_PASSWORD=yourpassword -p 6432:5432 -d postgres:latest

echo "Postgresql starting..."
sleep 3

# Veritabanı oluştur
docker exec -it postgress-test psql -U postgress -d postgress -c "CREATE DATABASE productapp"
sleep 3
echo  "Database productapp created"

# Tablo oluştur
docker exec -it postgress-test psql -U postgress -d productapp -c "
create table if not exists productapp
(id bigserial not null primary key,
 name varchar(255) not null,
 price double precision not null,
 discount double precision,
 store varchar(255) not null
);
"

sleep 3
echo "Table products created"
