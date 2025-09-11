
# Сервис подписок

Сервис управления онлайн-подписками пользователей. Написан на Go, использует PostgreSQL, Gin, GORM, Swagger и Docker.

## Особенности конфигурации

Нужно скопировать файл `.env.dev` в `.env`

Файл `.env.example` используется только как пример для Docker Compose.

## Makefile команды
###  Команды для локальной разработки
- `make test`            - Запуск интеграционных тестов
- `make coverage`        - Отчет о покрытии кода тестами
- `swagger-update`       - Обновление Swagger документации

### Работа с Docker
- `make up`              - Поднять docker-compose
- `make down`            - Остановить все контейнеры
- `make rebuild`         - Пересобрать все контейнеры
- `make restart`         - Рестартануть все контейнеры
- `make logs`            - Просмотр логов API
- `make build-prod TAG=0.1.8`   - Сборка прод-образа с нужным тегом
- `make push-prod TAG=0.1.8`    - Публикация прод-образа с нужным тегом
- `make run-prod TAG=0.1.8`     - Запуск прод-образа с нужным тегом локально"

### Миграции
- `make migrate-up`      - Накатить все миграции
- `make migrate-down`    - Откатить все миграции
- `make migrate-force v=3`  - Проставить версию миграции
- `make migrate-goto v=5`   - Перейти к миграции №5
    

## Как запустить дев окружение
1.  Скопировать файл `.env.dev` в `.env`
2. `make up`

Сервис будет доступен на http://localhost:8080/

http://localhost:8080/swagger/ - Swagger документация

http://localhost:8080/subscriptions - Ручки проекта

[/internal/swagger/swagger.yaml](/internal/swagger/swagger.yaml)

[/internal/swagger/swagger.json](/internal/swagger/swagger.json)

## Как запустить собранный прод-образ локально

```
make run-prod TAG=0.0.1
```
Запустит прод образ на порту :8081.
Обращения будут идти в дев базу

### Запуск тестов

В проекте представлены только интеграционные тесты, так как логика микросервиса
не подразумевает сложное деление кода и перепроверка каждого изолированного слоя
избыточна.

Интеграционные тесты используют базу проекта с постфиксом `_test`

- Интеграционные тесты:
```bash
make test
```

-  Отчёт покрытия:
```bash
make coverage
```

## Архитектура

Проект построен по [project-layout](https://github.com/golang-standards/project-layout).

Внутри `internal/` используется разделение по слоям:

- handlers — HTTP-ручки
- repository — работа с БД
- models — модели данных
- database — подключение к БД
- config — конфигурация через переменные окружения
- swagger — документация API
```
subscriptions-service/
├── cmd/
│   └── subscriptions-api/
│       └── main.go                                    - Основной код приложения
├── internal/
│   ├── handlers/
│   │   └── handlers.go                                - Хандлеры
│   ├── repository/
│   │   └── repository.go                              - Работа с базой
│   └── swagger/                                       - Автогенерируемая документация для Swagger
├── migrations/
│   ├── 0001_init.up.sql
│   └── 0001_init.down.sql
├── tests/
│   └── integration/
│       └── api_test.go                                - Интеграционные тесты
├── .air.toml                                          - Конфигурация Air для авторестарта демона в дев
├── .env
├── .env.dev                                           - .env для разработки (нужно скопировать в .env)
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile                                         - Dockerfile для прода
├── Dockerfile.dev                                     - Dockerfile для среды разработки
├── Dockerfile.test                                    - Dockerfile для интеграционных тестов
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── Subscriptions Service API.postman_collection.json  - Готовая коллекция запросов для Postman
```