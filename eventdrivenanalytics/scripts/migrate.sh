#!/bin/bash 
migrate \
  -path database/oltp/migrations \
  -database "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable" \
  up