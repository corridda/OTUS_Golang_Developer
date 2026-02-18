package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config содержит настройки подключения к MongoDB
type Config struct {
	URI             string
	Database        string
	ConnectTimeout  time.Duration
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		URI:             "mongodb://localhost:27017",
		Database:        "practice",
		ConnectTimeout:  10 * time.Second,
		MaxPoolSize:     100,
		MinPoolSize:     10,
		MaxConnIdleTime: 30 * time.Second,
	}
}

// Client обёртка над MongoDB клиентом
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

// NewClient создаёт новое подключение к MongoDB
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Проверяем соединение
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
	}, nil
}

// Close закрывает соединение
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Database возвращает базу данных
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection возвращает коллекцию
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Raw возвращает базовый клиент
func (c *Client) Raw() *mongo.Client {
	return c.client
}

// =============================================================================
// Базовые CRUD операции
// =============================================================================

// InsertOne вставляет один документ
func (c *Client) InsertOne(ctx context.Context, collection string, document interface{}) (primitive.ObjectID, error) {
	result, err := c.Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to insert document: %w", err)
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// InsertMany вставляет несколько документов
func (c *Client) InsertMany(ctx context.Context, collection string, documents []interface{}) ([]primitive.ObjectID, error) {
	result, err := c.Collection(collection).InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}

	ids := make([]primitive.ObjectID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		ids[i] = id.(primitive.ObjectID)
	}
	return ids, nil
}

// FindOne находит один документ
func (c *Client) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	err := c.Collection(collection).FindOne(ctx, filter).Decode(result)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("document not found")
	}
	return err
}

// FindByID находит документ по ID
func (c *Client) FindByID(ctx context.Context, collection string, id primitive.ObjectID, result interface{}) error {
	return c.FindOne(ctx, collection, bson.M{"_id": id}, result)
}

// Find находит несколько документов
func (c *Client) Find(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return c.Collection(collection).Find(ctx, filter, opts...)
}

// FindAll находит все документы и возвращает их в slice
func (c *Client) FindAll(ctx context.Context, collection string, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := c.Find(ctx, collection, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// UpdateOne обновляет один документ
func (c *Client) UpdateOne(ctx context.Context, collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	return c.Collection(collection).UpdateOne(ctx, filter, update)
}

// UpdateByID обновляет документ по ID
func (c *Client) UpdateByID(ctx context.Context, collection string, id primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error) {
	return c.UpdateOne(ctx, collection, bson.M{"_id": id}, update)
}

// UpdateMany обновляет несколько документов
func (c *Client) UpdateMany(ctx context.Context, collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	return c.Collection(collection).UpdateMany(ctx, filter, update)
}

// ReplaceOne заменяет один документ
func (c *Client) ReplaceOne(ctx context.Context, collection string, filter, replacement interface{}) (*mongo.UpdateResult, error) {
	return c.Collection(collection).ReplaceOne(ctx, filter, replacement)
}

// DeleteOne удаляет один документ
func (c *Client) DeleteOne(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	return c.Collection(collection).DeleteOne(ctx, filter)
}

// DeleteByID удаляет документ по ID
func (c *Client) DeleteByID(ctx context.Context, collection string, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return c.DeleteOne(ctx, collection, bson.M{"_id": id})
}

// DeleteMany удаляет несколько документов
func (c *Client) DeleteMany(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	return c.Collection(collection).DeleteMany(ctx, filter)
}

// Count считает документы
func (c *Client) Count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return c.Collection(collection).CountDocuments(ctx, filter)
}

// Exists проверяет существование документа
func (c *Client) Exists(ctx context.Context, collection string, filter interface{}) (bool, error) {
	count, err := c.Count(ctx, collection, filter)
	return count > 0, err
}

// =============================================================================
// Индексы
// =============================================================================

// CreateIndex создаёт индекс
func (c *Client) CreateIndex(ctx context.Context, collection string, keys interface{}, opts *options.IndexOptions) (string, error) {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}
	return c.Collection(collection).Indexes().CreateOne(ctx, indexModel)
}

// CreateUniqueIndex создаёт уникальный индекс
func (c *Client) CreateUniqueIndex(ctx context.Context, collection string, keys interface{}) (string, error) {
	return c.CreateIndex(ctx, collection, keys, options.Index().SetUnique(true))
}

// CreateTextIndex создаёт текстовый индекс
func (c *Client) CreateTextIndex(ctx context.Context, collection string, field string) (string, error) {
	return c.CreateIndex(ctx, collection, bson.D{{Key: field, Value: "text"}}, nil)
}

// CreateTTLIndex создаёт TTL индекс
func (c *Client) CreateTTLIndex(ctx context.Context, collection string, field string, expireAfter time.Duration) (string, error) {
	return c.CreateIndex(ctx, collection,
		bson.D{{Key: field, Value: 1}},
		options.Index().SetExpireAfterSeconds(int32(expireAfter.Seconds())),
	)
}

// ListIndexes возвращает список индексов
func (c *Client) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	cursor, err := c.Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

// DropIndex удаляет индекс
func (c *Client) DropIndex(ctx context.Context, collection, name string) error {
	_, err := c.Collection(collection).Indexes().DropOne(ctx, name)
	return err
}

// =============================================================================
// Транзакции
// =============================================================================

// WithTransaction выполняет функцию в транзакции
func (c *Client) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := c.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkWrite выполняет bulk операции
func (c *Client) BulkWrite(ctx context.Context, collection string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return c.Collection(collection).BulkWrite(ctx, models, opts...)
}

// BulkInsert выполняет массовую вставку
func (c *Client) BulkInsert(ctx context.Context, collection string, documents []interface{}) (*mongo.BulkWriteResult, error) {
	models := make([]mongo.WriteModel, len(documents))
	for i, doc := range documents {
		models[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}
	return c.BulkWrite(ctx, collection, models, options.BulkWrite().SetOrdered(false))
}
