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

// NoteRepository репозиторий заметок
type NoteRepository struct {
	client     *Client
	collection string
}

// NewTaskRepository создаёт новый репозиторий
func NewNoteRepository(client *Client) *NoteRepository {
	return &NoteRepository{
		client:     client,
		collection: "notes",
	}
}

// Create создаёт заметку
func (r *NoteRepository) Create(ctx context.Context, note *model.Note) error {
	id, err := r.client.InsertOne(ctx, r.collection, note)
	if err != nil {
		return fmt.Errorf("ошибка создания новой заметки : %w", err)
	}
	note.Id = id
	return nil
}

// GetAll считывает все заметки из БД
func (r *NoteRepository) GetAll(ctx context.Context) ([]*model.Note, int64, error) {
	total, err := r.client.Count(ctx, r.collection, bson.D{})
	if err != nil {
		return nil, 0, err
	}

	var notes []*model.Note
	if err := r.client.FindAll(ctx, r.collection, bson.D{}, &notes, nil); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

// GetById считывает задачу из БД по ее Id
func (r *NoteRepository) GetById(ctx context.Context, id primitive.ObjectID) (*model.Note, error) {
	var note model.Note
	if err := r.client.FindByID(ctx, r.collection, id, &note); err != nil {
		return nil, fmt.Errorf("ошибка считывания заметки из БД по Id : %w", err)
	}
	return &note, nil
}

// UpdateById обновляет документ заметки в БД по Id
func (r *NoteRepository) UpdateById(
	ctx context.Context,
	id primitive.ObjectID,
	name string,
	description string,
	alarmTimeStamp string,
) error {
	t, err := time.Parse("02.01.2006 15:04", alarmTimeStamp)
	if err != nil {
		return fmt.Errorf("Неверный формат даты/времени: %w", err)
	}
	alarmTimeStamp = t.Format(time.RFC3339)
	update := bson.M{
		"$set": bson.M{
			"name":           name,
			"description":    description,
			"alarmTimeStamp": alarmTimeStamp,
		},
	}

	result, err := r.client.UpdateByID(ctx, r.collection, id, update)
	if err != nil {
		return fmt.Errorf("Ошибка обновления заметки: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("Заметка не найдена.")
	}
	return nil
}

// DeleteById удаляет документ заметки в БД по Id
func (r *NoteRepository) DeleteById(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	result, err := r.client.DeleteByID(ctx, r.collection, id)
	if err != nil {
		return fmt.Errorf("Ошибка удаления заметки: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("Заметка не найдена.")
	}
	return nil
}

// EnsureIndexes создаёт необходимые индексы
func (r *NoteRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		keys interface{}
		opts *options.IndexOptions
	}{
		{bson.D{{Key: "name", Value: "text"}}, nil},
		{bson.D{{Key: "alarmTimeStamp", Value: 1}}, nil},
	}

	for _, idx := range indexes {
		if _, err := r.client.CreateIndex(ctx, r.collection, idx.keys, idx.opts); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}
