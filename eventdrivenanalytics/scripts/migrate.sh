#!/bin/bash 
migrate \
  -path database/oltp/migrations \
  -database "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable" \
  up

migrate \
  -path database/olap \
  -database "postgres://postgres:postgres@localhost:5433/analytics?sslmode=disable" \
  up