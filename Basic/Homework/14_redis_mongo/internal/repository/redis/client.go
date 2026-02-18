package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config содержит настройки подключения к Redis
type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Client обёртка над Redis клиентом с дополнительными методами
type Client struct {
	rdb *redis.Client
}

// NewClient создаёт новое подключение к Redis
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close закрывает соединение с Redis
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Raw возвращает базовый клиент для прямого доступа
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// =============================================================================
// Базовые операции со строками
// =============================================================================

// Set устанавливает значение с опциональным TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get получает строковое значение
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("key %s not found", key)
	}
	return val, err
}

// GetOrSet получает значение или устанавливает его через callback
func (c *Client) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, redis.Nil) {
		return "", err
	}

	// Ключ не найден - вычисляем и сохраняем
	val, err = fn()
	if err != nil {
		return "", fmt.Errorf("callback error: %w", err)
	}

	if err := c.rdb.Set(ctx, key, val, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to set value: %w", err)
	}

	return val, nil
}

// SetJSON сериализует объект в JSON и сохраняет
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// GetJSON получает значение и десериализует из JSON
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("key %s not found", key)
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// =============================================================================
// Атомарные счётчики
// =============================================================================

// Incr увеличивает счётчик на 1
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// IncrBy увеличивает счётчик на указанное значение
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

// Decr уменьшает счётчик на 1
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, key).Result()
}

// =============================================================================
// Работа с хэш-таблицами
// =============================================================================

// HSet устанавливает поля хэш-таблицы
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

// HGet получает значение поля
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.rdb.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("field %s not found in hash %s", field, key)
	}
	return val, err
}

// HGetAll получает все поля хэш-таблицы
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// HIncrBy увеличивает числовое поле хэш-таблицы
func (c *Client) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return c.rdb.HIncrBy(ctx, key, field, incr).Result()
}

// =============================================================================
// Вспомогательные методы
// =============================================================================

// Exists проверяет существование ключей
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Del удаляет ключи
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Expire устанавливает TTL для ключа
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// TTL возвращает оставшееся время жизни ключа
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Keys возвращает ключи по паттерну (использовать осторожно в production!)
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, pattern).Result()
}

// Scan итеративно получает ключи по паттерну (безопасно для production)
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.rdb.Scan(ctx, cursor, match, count).Result()
}

// FlushDB очищает текущую базу данных
func (c *Client) FlushDB(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}
