# Запуск контейнера
````
docker compose up -d
````

# Проверка статуса
````
docker compose ps
````

# Ессли нужно зайти в CLI
```
docker exec -it hw15-postgres psql -Uotus_user -dremindables
```

# Создаем миграции
```
go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir migrations create init sql
```

# Применяем миграции и проверяем
```
goose -dir migrations postgres "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable" up

goose -dir migrations postgres "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable" status

goose -dir migrations postgres "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable" reset

goose -dir migrations postgres "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable" down

goose -dir migrations postgres "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable" down-to 001
```
