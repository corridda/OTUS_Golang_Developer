package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/15_postgresql/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Remindable interface {
	String() string
	ChangeAlarm(string)
}

type RemindableId struct {
	Id int `form:"id" binding:"required"`
}

type NewTask struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	DueDate     string `json:"dueDate" binding:"required"`
}

type ChangingTask struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
}

type NewNote struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description" binding:"required"`
	AlarmTimeStamp string `json:"alarmTime" binding:"required"`
}

type ChangingNote struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	AlarmTimeStamp string `json:"alarmTimeStamp"`
}

// CreateNewRemindable Создать объект типа, реализующего Remindable
func CreateNewRemindable(
	ctx context.Context,
	db *sql.DB,
	name,
	descr,
	futurePoint string,
	isTask bool,
) error {
	var remindable Remindable
	var rows_count sql.NullInt64
	if isTask {
		err := db.QueryRowContext(ctx, "SELECT MAX((task #>> '{id}')::int) from tasks").Scan(&rows_count)
		if err != nil {
			return fmt.Errorf("ошибка считывания количества строк из БД: %v", err)
		}
		task, err := model.NewTask(rows_count, name, descr, futurePoint)
		if err != nil {
			return err
		}
		remindable = &task
	} else {
		err := db.QueryRowContext(ctx, "SELECT MAX((note #>> '{id}')::int) from notes").Scan(&rows_count)
		if err != nil {
			return fmt.Errorf("ошибка считывания количества строк из БД: %v", err)
		}
		note, err := model.NewNote(rows_count, name, descr, futurePoint)
		if err != nil {
			return err
		}
		remindable = &note
	}
	if err := SaveRemindable(ctx, db, remindable); err != nil {
		return err
	}
	return nil
}

// Обработка ошибки вставки в БД с ограничением на уникальность
func handleUnique(value any, err error) error {
	var pgErr *pgconn.PgError
	var s1, s2 string
	switch v := value.(type) {
	case *model.Task:
		s1 = fmt.Sprintf("Ошибка создания/изменения задачи - задача с именем '%s' уже существует", v.Name)
		s2 = "Ошибка создания/изменения задачи"
	case *model.Note:
		s1 = fmt.Sprintf("Ошибка создания/изменения заметки - заметка с именем '%s' уже существует", v.Name)
		s2 = "Ошибка создания/изменения заметки"
	}
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			// Обработка ошибки вставки в поле с ограничением на уникальность
			return fmt.Errorf("%s", s1)
		}
	}
	// Обработка других типов ошибок
	return fmt.Errorf("%s", s2)
}

// SaveRemindable Сохранить объект типа, реализующего Remindable в БД PostgreSQL
func SaveRemindable(
	ctx context.Context,
	db *sql.DB,
	remindable Remindable,
) error {
	r := remindable
	var rows int64
	var result sql.Result

	switch value := r.(type) {
	case *model.Task:
		task, err := json.Marshal(value)
		if err != nil {
			return err
		}
		result, err = db.ExecContext(
			ctx,
			`INSERT INTO tasks(task) VALUES($1)`,
			task,
		)
		if err != nil {
			err = handleUnique(value, err)
			if err != nil {
				return err
			}
		}
	case *model.Note:
		note, err := json.Marshal(value)
		if err != nil {
			return err
		}
		result, err = db.ExecContext(
			ctx,
			`INSERT INTO notes(note) VALUES($1)`,
			note,
		)
		if err != nil {
			err = handleUnique(value, err)
			if err != nil {
				return err
			}
		}
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return err
	}
	return nil
}

// GetTasks
// @Summary Получить все задачи
// @Tags Задачи
// @Produce	json
// @Success 200 {string} string "Getting tasks is successful"
// @Failure 500 {string} string "Getting notes failed due to internal server error"
// @Router /api/tasks/items [get]
// Обработка Get-запроса типа /api/items для задач
func GetTasks(ctx context.Context, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tasks := make([]model.Task, 0)
		rows, err := db.QueryContext(ctx, "SELECT * FROM tasks")
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				log.Fatalf("Ошибка закрытия объекта sql.Rows: %v\n", err)
			}
		}(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"InternalServerError": "Ошибка считывания задач из БД"})
			return
		}

		for rows.Next() {
			var task model.Task
			var id int64
			var taskByte []byte
			var createdAt time.Time
			var updatedAt sql.NullTime
			if err := rows.Scan(&id, &taskByte, &createdAt, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"InternalServerError": err.Error()})
				return
			}
			if err = json.Unmarshal(taskByte, &task); err != nil {
				log.Fatalf("Ошибка десериализации JSON: %s\n", err)
			}
			tasks = append(tasks, task)
		}
		c.JSON(http.StatusOK, tasks)
	}
}

// GetTasksById
// @Summary Получить задачу по ее идентификатору
// @Tags Задача по ID
// @Produce	json
// @Param id query int true "Task ID"
// @Success 200 {string} string "Getting the task is successful"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found: such a task doesn't exist"
// @Router /api/tasks/item/id [get]
// Обработка Get-запрос типа /api/item/id для задач
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/tasks/item/id?id=1
func GetTasksById(ctx context.Context, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		var task model.Task
		var taskByte []byte
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT task FROM tasks WHERE (task #>> '{id}')::int=$1",
				taskId.Id,
			).Scan(&taskByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(taskByte, &task); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				c.JSON(http.StatusOK, task)
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID задачи"})
			return
		}
	}
}

// GetNotes
// @Summary Получить все заметки
// @Tags Заметки
// @Produce	json
// @Success 200 {string} string "Getting notes is successful"
// @Failure 500 {string} string "Getting notes failed due to internal server error"
// @Router /api/notes/items [get]
// Обработка Get-запроса типа /api/items для заметок
func GetNotes(ctx context.Context, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		notes := make([]model.Note, 0)
		rows, err := db.QueryContext(ctx, "SELECT * FROM notes")
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				log.Fatalf("Ошибка закрытия объекта sql.Rows: %v\n", err)
			}
		}(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"InternalServerError": "Ошибка считывания заметок из БД"})
			return
		}

		for rows.Next() {
			var note model.Note
			var id int64
			var noteByte []byte
			var createdAt time.Time
			var updatedAt sql.NullTime
			if err := rows.Scan(&id, &noteByte, &createdAt, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"InternalServerError": err.Error()})
				return
			}
			if err = json.Unmarshal(noteByte, &note); err != nil {
				log.Fatalf("Ошибка десериализации JSON: %s\n", err)
			}
			notes = append(notes, note)
		}
		c.JSON(http.StatusOK, notes)
	}
}

// GetNotesById
// @Summary Получить заметку по ее идентификатору
// @Tags Заметка по ID
// @Produce	json
// @Param id query int true "Note ID"
// @Success 200 {string} string "Getting the note is successful"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found: such a note doesn't exist"
// @Router /api/notes/item/id [get]
// Обработка Get-запроса типа /api/item/id для заметок
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/notes/item/id?id=1
func GetNotesById(ctx context.Context, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		var note model.Note
		var noteByte []byte
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT note FROM notes WHERE (note #>> '{id}')::int=$1",
				noteId.Id,
			).Scan(&noteByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", noteId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(noteByte, &note); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				c.JSON(http.StatusOK, note)
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID заметки"})
			return
		}
	}
}

// PostNewTask
// @Summary Создать новую задачу
// @Tags Новая задача
// @Accept	json
// @Produce	json
// @Param newTask body NewTask true "Task data" body is the new task attributes
// @Success 200 {string} string "The task has been successfully created"
// @Failure 400 {string} string "Invalid request: the task hasn't been created"
// @Router /api/tasks/item [post]
// Обработка Post-запроса типа /api/item для задач
func PostNewTask(
	ctx context.Context,
	db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		newTask := NewTask{}
		err := c.ShouldBindJSON(&newTask)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = CreateNewRemindable(
			ctx,
			db,
			newTask.Name,
			newTask.Description,
			newTask.DueDate,
			true,
		)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"OK": "Создана новая задача"})
		result, err := db.ExecContext(
			ctx,
			`INSERT INTO remindables_log(description) VALUES($1)`,
			fmt.Sprintf("Создана новая задача %v", newTask.Name),
		)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"BadRequest": "Ошибка логирования в PostgreSQL создания новой задачи"},
			)
			return
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != 1 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"BadRequest": "Ошибка логирования в PostgreSQL создания новой задачи"},
			)
			return
		}
	}
}

// PostNewNote
// @Summary Создать новую заметку
// @Tags Новая заметка
// @Accept	json
// @Produce	json
// @Param newNote body NewNote true "Note data" body is the new note attributes
// @Success 200 {string} string "The note has been successfully created"
// @Failure 400 {string} string "Invalid request: the note hasn't been created"
// @Router /api/notes/item [post]
// Обработка Post-запроса типа /api/item для заметок
func PostNewNote(
	ctx context.Context,
	db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		newNote := NewNote{}
		err := c.ShouldBindJSON(&newNote)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = CreateNewRemindable(
			ctx,
			db,
			newNote.Name,
			newNote.Description,
			newNote.AlarmTimeStamp,
			false,
		)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"OK": "Создана новая заметка"})
		result, err := db.ExecContext(
			ctx,
			`INSERT INTO remindables_log(description) VALUES($1)`,
			fmt.Sprintf("Создана новая заметка %v", newNote.Name),
		)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"BadRequest": "Ошибка логирования в PostgreSQL создания новой заметки"},
			)
			return
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != 1 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"BadRequest": "Ошибка логирования в PostgreSQL создания новой заметки"},
			)
			return
		}
	}
}

// PutTaskById
// @Summary Изменить задачу по ее ID
// @Tags Изменить задачу
// @Accept	json
// @Produce	json
// @Param id query int true "Task ID"
// @Param updatedTask body ChangingTask true "Task data" body is the updating task attributes
// @Success 200 {string} string "The task has been successfully updated"
// @Failure 400 {string} string "Invalid request: the task hasn't been updated"
// @Router /api/tasks/item/id [put]
// Обработка Put-запроса типа /api/item/id для задач
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/tasks/item/id?id=1
func PutTaskById(
	ctx context.Context,
	db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		var task model.Task
		var taskByte []byte
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT task FROM tasks WHERE (task #>> '{id}')::int=$1",
				taskId.Id,
			).Scan(&taskByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(taskByte, &task); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				changingTask := ChangingTask{}
				err = c.ShouldBindJSON(&changingTask)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				task.Name = changingTask.Name
				task.Description = changingTask.Description
				newDueDate, err := time.Parse("02.01.2006", changingTask.DueDate)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Введенная дата исполнения имеет не корректный формат."})
					return
				}
				task.DueDate = newDueDate
				task.Status = model.Updated

				taskJSON, err := json.Marshal(task)
				if err != nil {
					log.Fatalf("Ошибка сериализации в JSON %s\n", err)
				}

				result, err := db.ExecContext(ctx, `
					UPDATE tasks
					SET task = $1, updated_at = now()
					WHERE (task #>> '{id}')::int = $2`,
					taskJSON, task.Id,
				)
				if err != nil {
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время обновления задачи с id=%d произошла ошибка: %v\n", task.Id, err.Error())})
						return
					}
					rows, err := result.RowsAffected()
					if err != nil || rows != 1 {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время обновления задачи с id=%d произошла ошибка: %v\n", task.Id, err.Error())})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"Изменена задача": task})
				result, err = db.ExecContext(
					ctx,
					`INSERT INTO remindables_log(description) VALUES($1)`,
					fmt.Sprintf("Изменена задача %v", task.Name),
				)
				if err != nil {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL изменения задачи"},
					)
					return
				}
				rows, err := result.RowsAffected()
				if err != nil || rows != 1 {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL изменения задачи"},
					)
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID задачи"})
			return
		}
	}
}

// PutNoteById
// @Summary Изменить заметку по ее ID
// @Tags Изменить заметку
// @Accept	json
// @Produce	json
// @Param id query int true "Note ID"
// @Param updatedNote body ChangingNote true "Note data" body is the updating note attributes
// @Success 200 {string} string "The note has been successfully updated"
// @Failure 400 {string} string "Invalid request: the note hasn't been updated"
// @Router /api/notes/item/id [put]
// Обработка Put-запроса типа /api/item/id для заметок
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/notes/item/id?id=1
func PutNoteById(
	ctx context.Context,
	db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		var note model.Note
		var noteByte []byte
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT note FROM notes WHERE (note #>> '{id}')::int=$1",
				noteId.Id,
			).Scan(&noteByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", noteId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(noteByte, &note); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				changingNote := ChangingNote{}
				err = c.ShouldBindJSON(&changingNote)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				note.Name = changingNote.Name
				note.Description = changingNote.Description
				newAlarmTimeStamp, err := time.Parse("02.01.2006 15:04", changingNote.AlarmTimeStamp)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Введенные дата-время напоминания имеют не корректный формат."})
					return
				}
				note.AlarmTimeStamp = newAlarmTimeStamp

				noteJSON, err := json.Marshal(note)
				if err != nil {
					log.Fatalf("Ошибка сериализации в JSON %s\n", err)
				}

				result, err := db.ExecContext(ctx, `
					UPDATE notes
					SET note = $1, updated_at = now()
					WHERE (note #>> '{id}')::int = $2`,
					noteJSON, note.Id,
				)
				if err != nil {
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время обновления заметки с id=%d произошла ошибка: %v\n", note.Id, err.Error())})
						return
					}
					rows, err := result.RowsAffected()
					if err != nil || rows != 1 {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время обновления заметки с id=%d произошла ошибка: %v\n", note.Id, err.Error())})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"Изменена заметка": note})
				result, err = db.ExecContext(
					ctx,
					`INSERT INTO remindables_log(description) VALUES($1)`,
					fmt.Sprintf("Изменена заметка %v", note.Name),
				)
				if err != nil {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL изменения заметка"},
					)
					return
				}
				rows, err := result.RowsAffected()
				if err != nil || rows != 1 {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL изменения заметка"},
					)
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID заметки"})
			return
		}
	}
}

// DeleteTaskById
// @Summary Удалить задачу по ее ID
// @Tags Удалить задачу
// @Produce	json
// @Param id query int true "Task ID"
// @Success 200 {string} string "The task has been successfully deleted"
// @Failure 400 {string} string "Invalid request: the task hasn't been deleted"
// @Failure 404 {string} string "Not found: such a task doesn't exist"
// @Router /api/tasks/item/id [delete]
// Обработка Delete-запроса типа /api/item/id для задач
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/tasks/item/id?id=1
func DeleteTaskById(ctx context.Context, db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		var task model.Task
		var taskByte []byte
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT task FROM tasks WHERE (task #>> '{id}')::int=$1",
				taskId.Id,
			).Scan(&taskByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(taskByte, &task); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				result, err := db.ExecContext(ctx, `
					DELETE from tasks
					WHERE (task #>> '{id}')::int = $1`,
					task.Id,
				)
				if err != nil {
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время удаления задачи с id=%d произошла ошибка: %v\n", task.Id, err.Error())})
						return
					}
					rows, err := result.RowsAffected()
					if err != nil || rows != 1 {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время удаления задачи с id=%d произошла ошибка: %v\n", task.Id, err.Error())})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"Удалена задача": task})
				result, err = db.ExecContext(
					ctx,
					`INSERT INTO remindables_log(description) VALUES($1)`,
					fmt.Sprintf("Удалена задача %v", task.Name),
				)
				if err != nil {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL удаления задачи"},
					)
					return
				}
				rows, err := result.RowsAffected()
				if err != nil || rows != 1 {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL удаления задачи"},
					)
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID задачи"})
			return
		}
	}
}

// DeleteNoteById
// @Summary Удалить заметку по ее ID
// @Tags Удалить заметку
// @Produce	json
// @Param id query int true "Note ID"
// @Success 200 {string} string "The note has been successfully deleted"
// @Failure 400 {string} string "Invalid request: the note hasn't been deleted"
// @Failure 404 {string} string "Not found: such a note doesn't exist"
// @Router /api/notes/item/id [delete]
// Обработка Delete-запроса типа /api/item/id для заметок
// id в запросе передается в виде целого неотрицательного числа, большего нуля, напр.:
// /api/notes/item/id?id=1
func DeleteNoteById(ctx context.Context, db *sql.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		var note model.Note
		var noteByte []byte
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			err := db.QueryRowContext(
				ctx,
				"SELECT note FROM notes WHERE (note #>> '{id}')::int=$1",
				noteId.Id,
			).Scan(&noteByte)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", noteId.Id)},
				)
				return
			case err != nil:
				log.Fatalf("Ошибка запроса: %v\n", err)
			default:
				if err = json.Unmarshal(noteByte, &note); err != nil {
					log.Fatalf("Ошибка десериализации JSON: %s\n", err)
				}
				result, err := db.ExecContext(ctx, `
					DELETE from notes
					WHERE (note #>> '{id}')::int = $1`,
					note.Id,
				)
				if err != nil {
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время удаления заметки с id=%d произошла ошибка: %v\n", note.Id, err.Error())})
						return
					}
					rows, err := result.RowsAffected()
					if err != nil || rows != 1 {
						c.JSON(http.StatusBadRequest, gin.H{"Ошибка": fmt.Sprintf("Во время удаления заметки с id=%d произошла ошибка: %v\n", note.Id, err.Error())})
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"Удалена заметка": note})
				result, err = db.ExecContext(
					ctx,
					`INSERT INTO remindables_log(description) VALUES($1)`,
					fmt.Sprintf("Удалена заметка %v", note.Name),
				)
				if err != nil {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL удаления заметки"},
					)
					return
				}
				rows, err := result.RowsAffected()
				if err != nil || rows != 1 {
					c.JSON(
						http.StatusBadRequest,
						gin.H{"BadRequest": "Ошибка логирования в PostgreSQL удаления заметки"},
					)
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Ошибка": "Некорректный ID заметки"})
			return
		}
	}
}
