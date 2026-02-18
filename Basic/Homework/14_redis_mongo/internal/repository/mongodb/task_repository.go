package mongodb

import (
	"context"
	"fmt"
	"time"

	//"fmt"
	//"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	//"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskRepository репозиторий задач
type TaskRepository struct {
	client     *Client
	collection string
}

// NewTaskRepository создаёт новый репозиторий
func NewTaskRepository(client *Client) *TaskRepository {
	return &TaskRepository{
		client:     client,
		collection: "tasks",
	}
}

// Create создаёт задачу
func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	id, err := r.client.InsertOne(ctx, r.collection, task)
	if err != nil {
		return fmt.Errorf("ошибка создания новой задачи : %w", err)
	}
	task.Id = id
	return nil
}

// GetAll считывает все задачи из БД
func (r *TaskRepository) GetAll(ctx context.Context) ([]*model.Task, int64, error) {
	total, err := r.client.Count(ctx, r.collection, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	var tasks []*model.Task
	if err := r.client.FindAll(ctx, r.collection, bson.D{}, &tasks, nil); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetById считывает задачу из БД по ее Id
func (r *TaskRepository) GetById(ctx context.Context, id primitive.ObjectID) (*model.Task, error) {
	var task model.Task
	if err := r.client.FindByID(ctx, r.collection, id, &task); err != nil {
		return nil, fmt.Errorf("ошибка считывания задачи из БД по Id : %w", err)
	}
	return &task, nil
}

// UpdateById обновляет документ задачи в БД по Id
func (r *TaskRepository) UpdateById(
	ctx context.Context,
	id primitive.ObjectID,
	name string,
	description string,
	dueDate string,
) error {
	t, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		return fmt.Errorf("Неверный формат даты: %w", err)
	}
	dueDate = t.Format(time.RFC3339)
	update := bson.M{
		"$set": bson.M{
			"name":        name,
			"description": description,
			"dueDate":     dueDate,
			"status":      model.Updated,
		},
	}

	result, err := r.client.UpdateByID(ctx, r.collection, id, update)
	if err != nil {
		return fmt.Errorf("Ошибка обновления задачи: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("Задача не найдена.")
	}
	return nil
}

// DeleteById удаляет документ задачи в БД по Id
func (r *TaskRepository) DeleteById(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	result, err := r.client.DeleteByID(ctx, r.collection, id)
	if err != nil {
		return fmt.Errorf("Ошибка удаления задачи: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("Задача не найдена.")
	}
	return nil
}

// EnsureIndexes создаёт необходимые индексы
func (r *TaskRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		keys interface{}
		opts *options.IndexOptions
	}{
		{bson.D{{Key: "name", Value: "text"}}, nil},
		{bson.D{{Key: "status", Value: "text"}}, nil},
		{bson.D{{Key: "dueDate", Value: 1}}, nil},
		{bson.D{{Key: "initTimeStamp", Value: -1}}, nil},
	}

	for _, idx := range indexes {
		if _, err := r.client.CreateIndex(ctx, r.collection, idx.keys, idx.opts); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}
