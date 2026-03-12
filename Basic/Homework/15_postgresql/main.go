package main

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/15_postgresql/internal/repository"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx := context.Background()

	// Подключаемся к PostgreSQL
	dsn := "host=localhost port=5432 user=otus_user password=otus_password dbname=remindables sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("connect db", "error", err)
		return
	}

	defer func() {
		err := db.Close()
		if err != nil {
			slog.Error("close db", "error", err)
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		slog.Error("ping db", "error", err)
		return
	}

	// Накатываем миграции
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("cannot set dialect", "error", err)
		return
	}

	const migrate = "./migrations"

	if err := goose.UpContext(ctx, db, migrate); err != nil {
		slog.Error("cannot do up migration", "error", err)
		return
	}

	// Создаём роутер
	r := gin.Default()
	api := r.Group("/api")
	apiTasks := api.Group("/tasks")
	apiNotes := api.Group("/notes")

	// Endpoints

	// /api/tasks/items
	apiTasks.GET("items", repository.GetTasks(ctx, db))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.GET("item/id", repository.GetTasksById(ctx, db))

	// /api/notes/items
	apiNotes.GET("items", repository.GetNotes(ctx, db))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.GET("item/id", repository.GetNotesById(ctx, db))

	// /api/tasks/item
	apiTasks.POST("item", repository.PostNewTask(ctx, db))

	// /api/notes/item
	apiNotes.POST("item", repository.PostNewNote(ctx, db))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.PUT("item/id", repository.PutTaskById(ctx, db))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.PUT("item/id", repository.PutNoteById(ctx, db))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.DELETE("item/id", repository.DeleteTaskById(ctx, db))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.DELETE("item/id", repository.DeleteNoteById(ctx, db))

	// Запуск сервера на :8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
