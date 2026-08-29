up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate:
	./scripts/migrate.sh

psql:
	psql postgres://postgres:postgres@localhost:5432/orders

seed:
	psql postgres://postgres:postgres@localhost:5432/orders \
	-f database/oltp/seed/seed.sql

analytics:
	docker exec -it olap-db psql -U postgres -d analytics