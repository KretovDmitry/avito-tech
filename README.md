# Go RESTful API for banner service


## Начало


```shell
# запуск постгреса в докере
make db-start

# запуск редиса в докере
make redis-start

# наполнить базу тестовыми данными
make testdata

# запуск
make run

# запуск с рестартом при любом изменение файлов проекта
# требуется fswatch
make run-live
```

Адрес `http://127.0.0.1:8080`. Эндпойнты:

* `GET /user_banner`: получение баннеров для пользователя
* `GET /banner`: получение всех баннеров c фильтрацией по фиче и/или тегу админом`
* `POST /banner`: создание баннера админом
* `DELETE /banner`: асинхронное удаление баннеров админом
* `PATCH /banner/:id`: обновление баннеров админом
* `DELETE /banner/:id`: синхронное удаление  баннера админом

## Запросы в Постмане

[<img src="https://run.pstmn.io/button.svg" alt="Run In Postman" style="width: 128px; height: 32px;">](https://god.gw.postman.com/run-collection/28228886-62014812-efec-4b31-b24f-b1ef98b40980?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D28228886-62014812-efec-4b31-b24f-b1ef98b40980%26entityType%3Dcollection%26workspaceId%3D8267f593-6a79-467b-8380-fc86774160f2)

## Структура проекта

 
```
.
├── cmd                  исполняемые файлы
│   └── server           
├── config               файлы конфигураций для разных сред
├── internal             приватные пакеты
│   ├── banner           сервис баннеров
│   ├── auth             аутентификация
│   ├── config           для загрузки конфига
│   ├── jwt              для работы с токеном
│   ├── user             пользователи
│   └── test             ... не успел
├── migrations           миграции бд
├── pkg                  публичные пакеты
│   ├── accesslog        логирование каждого запроса
│   ├── log              логгер
└── testdata             скрипт для наполнения бд тестовыми данными
```

### Все доступные команды make

```shell
build                          build the API server binary
build-docker                   build the API server as a docker image
clean                          remove temporary files
db-start                       start the database server
db-stop                        stop the database server
fmt                            run "go fmt" on all Go packages
help                           help information about make commands
lint                           run golangchi lint on all Go package
migrate-down                   revert database to the last migration step
migrate-new                    create a new database migration
migrate-reset                  reset database and re-run all migrations
migrate                        run all new database migrations
redis-start                    start the redis server
redis-stop                     stop the redis server
run-live                       run the API server with live reload support (requires fswatch)
run-restart                    restart the API server
run                            run the API server
sqlc-generate                  generate Go code that presents type-safe interfaces to service queries
sqlc-verify                    verify schema changes
sqlc-vet                       run query analyzer on cloud hosted database
test-cover                     run unit tests and show test coverage information
testdata                       populate the database with test data
test                           run unit tests
version                        display the version of the API server
```

## Трудности

Не смог покрыть тестами из-за отсутствия времени среди рабочей недели. 
Проект написан за два дня.

Реализация получения всех баннеров админом далась с трудом из-за вариабельности
предоставляемых данных. Этот хэндлер совсем не DRY...