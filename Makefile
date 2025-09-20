DOCKER_IMAGE := subscriptions-service
.PHONY: $(MAKECMDGOALS)

%:
	@:

# =====================
# ПОМОЩЬ
# =====================
help:
	@echo "===== Команды для локальной разработки ====="
	@echo "make test            - Запуск интеграционных тестов"
	@echo "make coverage        - Отчет о покрытии кода тестами"
	@echo "swagger-update       - Обновление Swagger документации"
	@echo ""
	@echo "===== Работа с Docker ====="
	@echo "make up              - Поднять docker-compose"
	@echo "make down            - Остановить все контейнеры"
	@echo "make rebuild         - Пересобрать все контейнеры"
	@echo "make restart         - Рестартануть все контейнеры"
	@echo "make logs            - Просмотр логов API"
	@echo "make build-prod TAG=0.1.8  - Сборка прод-образа с нужным тегом"
	@echo "make push-prod TAG=0.1.8   - Публикация прод-образа с нужным тегом"
	@echo "make run-prod TAG=0.1.8    - Запуск прод-образа с нужным тегом локально"
	@echo ""
	@echo "===== Миграции ====="
	@echo "make create-migrations-file [FILENAME] - Создать файл миграции"
	@echo "make migrate-up      - Накатить все миграции"
	@echo "make migrate-down    - Откатить последнюю миграцию"
	@echo "make migrate-force v=3  - Проставить версию миграции"
	@echo "make migrate-goto v=5   - Перейти к миграции №5"


# =====================
# ТЕСТЫ И КАЧЕСТВО КОДА
# =====================
test:
	@echo ">>> Запуск интеграционных тестов..."
	@docker compose run --rm app-test go test ./tests/...

coverage:
	@echo ">>> Подсчет покрытия кода тестами..."
	@docker compose run --rm app-test go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

swagger-update:
	@echo ">>> Обновление swagger документации..."
	@docker compose exec app swag init -g ./cmd/subscriptions-api/main.go -o ./internal/swagger

# =====================
# DOCKER & COMPOSE
# =====================
up:
	docker compose up -d

down:
	docker compose down

rebuild:
	docker compose up -d --build

restart:
	docker compose down && docker compose up -d

logs:
	docker compose logs -f app

build-prod:
	@echo ">>> Сборка прод образа..."
	docker build -f Dockerfile -t $(DOCKER_IMAGE):$(TAG) .

push-prod:
	@echo ">>> Заливка прод образа..."
	docker push $(DOCKER_IMAGE):$(TAG)

run-prod:
	@echo ">>> Запуск прод образа локально на порту 8081 ..."
		docker run \
		-p 8081:8080 \
		--network subscriptions-service_default \
		-e DB_HOST=subscriptions-service-db \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=postgres \
		-e DB_NAME=subscriptions \
		-e SERVER_ADDRESS=:8080 \
		-e LOG_LEVEL=debug \
		$(DOCKER_IMAGE):$(TAG)

# =====================
# МИГРАЦИИ
# =====================
create-migrations-file:
	docker compose run --rm migrate create -dir /migrations -ext sql $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down 1

migrate-force:
	docker compose run --rm migrate force $(v)

migrate-goto:
	docker compose run --rm migrate goto $(v)
