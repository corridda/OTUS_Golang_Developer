package main

import (
	"context"
	"log"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/repository/mongodb"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/repository/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// Создаём клиент MongoDB
	cfg := mongodb.DefaultConfig()
	cfg.URI = "mongodb://admin:password@localhost:27017"
	cfg.Database = "remindables"

	client, err := mongodb.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Close(ctx)

	// Создаём репозитории
	taskRepo := mongodb.NewTaskRepository(client)
	noteRepo := mongodb.NewNoteRepository(client)

	// Создаём индексы
	if err := taskRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("Warning: failed to create user indexes: %v", err)
	}
	if err := noteRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("Warning: failed to create order indexes: %v", err)
	}

	// Создаём клиент Redis для логирования
	cfg_redis := redis.DefaultConfig()
	client_redis, err := redis.NewClient(cfg_redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer client_redis.Close()

	// Создаём роутер
	r := gin.Default()
	api := r.Group("/api")
	apiTasks := api.Group("/tasks")
	apiNotes := api.Group("/notes")

	// Endpoints

	// /api/tasks/items
	apiTasks.GET("items", repository.GetTasks(ctx, taskRepo))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.GET("item/id", repository.GetTasksById(ctx, taskRepo))

	// /api/notes/items
	apiNotes.GET("items", repository.GetNotes(ctx, noteRepo))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.GET("item/id", repository.GetNotesById(ctx, noteRepo))

	// /api/tasks/item
	apiTasks.POST("item", repository.PostNewTask(
		ctx,
		taskRepo,
		noteRepo,
		client_redis,
	))

	// /api/notes/item
	apiNotes.POST("item", repository.PostNewNote(
		ctx,
		taskRepo,
		noteRepo,
		client_redis,
	))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.PUT("item/id", repository.PutTaskById(
		ctx,
		taskRepo,
		client_redis,
	))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.PUT("item/id", repository.PutNoteById(
		ctx,
		noteRepo,
		client_redis,
	))

	// /api/tasks/item/id/?id=<id_integer_number>
	apiTasks.DELETE("item/id", repository.DeleteTaskById(
		ctx,
		taskRepo,
		client_redis,
	))

	// /api/notes/item/id/?id=<id_integer_number>
	apiNotes.DELETE("item/id", repository.DeleteNoteById(
		ctx,
		noteRepo,
		client_redis,
	))

	// Запуск сервера на :8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}

}
