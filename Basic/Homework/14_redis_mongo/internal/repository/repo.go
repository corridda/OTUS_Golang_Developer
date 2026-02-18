package repository

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/model"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/repository/mongodb"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/14_redis_mongo/internal/repository/redis"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Remindable interface {
	String() string
	ChangeAlarm(string)
}

type RemindableId struct {
	Id string `form:"id" binding:"required"`
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
	taskRepo *mongodb.TaskRepository,
	noteRepo *mongodb.NoteRepository,
	name,
	descr,
	futurePoint string,
	isTask bool,
) error {
	var remindable Remindable
	if isTask {
		task := model.NewTask(name, descr, futurePoint)
		remindable = &task
	} else {
		note := model.NewNote(name, descr, futurePoint)
		remindable = &note
	}
	if err := SaveRemindable(ctx, taskRepo, noteRepo, remindable); err != nil {
		return err
	}
	return nil
}

// SaveRemindable Сохранить объект типа, реализующего Remindable в соотв. срезе и БД MongoDB
func SaveRemindable(
	ctx context.Context,
	taskRepo *mongodb.TaskRepository,
	noteRepo *mongodb.NoteRepository,
	remindable Remindable,
) error {
	r := remindable
	switch value := r.(type) {
	case *model.Task:
		if err := taskRepo.Create(ctx, value); err != nil {
			return err
		}
	case *model.Note:
		if err := noteRepo.Create(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

// GetTasks Обработка Get-запроса типа /api/items для задач
func GetTasks(
	ctx context.Context,
	taskRepo *mongodb.TaskRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		tasks, _, err := taskRepo.GetAll(ctx)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка считывания задач из БД"})
			return
		}
		c.JSON(http.StatusOK, tasks)
	}
}

// GetTasksById Обработка Get-запрос типа /api/item/id для задач
// id в запросе передается в виде hex-строки, напр.:
// /api/tasks/item/id?id=69959fd9aece410dd54f5739
func GetTasksById(ctx context.Context,
	taskRepo *mongodb.TaskRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			fmt.Printf("taskId: %v\n", taskId.Id)
			objectID, err := primitive.ObjectIDFromHex(taskId.Id)
			fmt.Printf("objectID: %v\n", objectID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID задачи"})
				return
			}
			task, err := taskRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%s не существует.", taskId.Id)},
				)
				return
			} else {
				c.JSON(http.StatusOK, task)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}

// GetNotes Обработка Get-запроса типа /api/items для заметок
func GetNotes(ctx context.Context,
	noteRepo *mongodb.NoteRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		notes, _, err := noteRepo.GetAll(ctx)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка считывания заметок из БД"})
			return
		}
		c.JSON(http.StatusOK, notes)
	}
}

// GetNotesById Обработка Get-запроса типа /api/item/id для заметок
// id в запросе передается в виде hex-строки, напр.:
// /api/notes/item/id?id=69959fd9aece410dd54f5739
func GetNotesById(ctx context.Context,
	noteRepo *mongodb.NoteRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			objectID, err := primitive.ObjectIDFromHex(noteId.Id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заметки"})
				return
			}
			note, err := noteRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%s не существует.", noteId.Id)},
				)
				return
			} else {
				c.JSON(http.StatusOK, note)
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}

// PostNewTask Обработка Post-запроса типа /api/item для задач
func PostNewTask(
	ctx context.Context,
	taskRepo *mongodb.TaskRepository,
	noteRepo *mongodb.NoteRepository,
	client_redis *redis.Client,
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
			taskRepo,
			noteRepo,
			newTask.Name,
			newTask.Description,
			newTask.DueDate,
			true,
		)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка создания новой задачи"})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"OK": "Создана новая задача"})
			t, err := time.Parse("02.01.2006", newTask.DueDate)
			err = client_redis.SetJSON(
				ctx,
				fmt.Sprintf("Создана новая задача %v", newTask.Name),
				newTask,
				time.Until(t),
			)
			if err != nil {
				panic(fmt.Errorf("Ошибка логирования в Redis новой задачи"))
			}
		}
	}
}

// PostNewNote Обработка Post-запроса типа /api/item для заметок
func PostNewNote(
	ctx context.Context,
	taskRepo *mongodb.TaskRepository,
	noteRepo *mongodb.NoteRepository,
	client_redis *redis.Client,
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
			taskRepo,
			noteRepo,
			newNote.Name,
			newNote.Description,
			newNote.AlarmTimeStamp,
			false,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка создания новой заметки"})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"OK": "Создана новая заметка"})
			t, err := time.Parse("02.01.2006 15:04", newNote.AlarmTimeStamp)
			err = client_redis.SetJSON(
				ctx,
				fmt.Sprintf("Создана новая заметка %v", newNote.Name),
				newNote,
				time.Until(t),
			)
			if err != nil {
				panic(fmt.Errorf("Ошибка логирования в Redis новой заметки"))
			}
		}
	}
}

// PutTaskById Обработка Put-запроса типа /api/item/id для задач
// id в запросе передается в виде hex-строки, напр.:
// /api/tasks/item/id/?id=69959fd9aece410dd54f5739
func PutTaskById(ctx context.Context,
	taskRepo *mongodb.TaskRepository,
	client_redis *redis.Client,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			objectID, err := primitive.ObjectIDFromHex(taskId.Id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID задачи"})
				return
			}

			changingTask := ChangingTask{}
			err = c.ShouldBindJSON(&changingTask)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			taskToBeChanged, err := taskRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%s не существует.", taskId.Id)},
				)
				return
			}

			err = taskRepo.UpdateById(
				ctx,
				objectID,
				changingTask.Name,
				changingTask.Description,
				changingTask.DueDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			changedTask, err := taskRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%s не существует.", taskId.Id)},
				)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Изменена задача": changedTask,
				})
				t, err := time.Parse("02.01.2006", changingTask.DueDate)
				err = client_redis.SetJSON(
					ctx,
					fmt.Sprintf("Изменена задача %v", taskToBeChanged.Name),
					changingTask,
					time.Until(t),
				)
				if err != nil {
					panic(fmt.Errorf("Ошибка логирования в Redis изменения задачи"))
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}

// PutNoteById Обработка Put-запроса типа /api/item/id для заметок
// id в запросе передается в виде hex-строки, напр.:
// /api/notes/item/id/?id=69959fd9aece410dd54f5739
func PutNoteById(ctx context.Context,
	noteRepo *mongodb.NoteRepository,
	client_redis *redis.Client,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			objectID, err := primitive.ObjectIDFromHex(noteId.Id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заметки"})
				return
			}

			changingNote := ChangingNote{}
			err = c.ShouldBindJSON(&changingNote)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			noteToBeDeleted, err := noteRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%s не существует.", noteId.Id)},
				)
				return
			}

			err = noteRepo.UpdateById(
				ctx,
				objectID,
				changingNote.Name,
				changingNote.Description,
				changingNote.AlarmTimeStamp)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			changedNote, err := noteRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Заметки с id=%s не существует.", noteId.Id)},
				)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Изменена заметка": changedNote,
				})
				t, err := time.Parse("02.01.2006 15:04", changingNote.AlarmTimeStamp)
				err = client_redis.SetJSON(
					ctx,
					fmt.Sprintf("Изменена заметка %v", noteToBeDeleted.Name),
					changingNote,
					time.Until(t),
				)
				if err != nil {
					panic(fmt.Errorf("Ошибка логирования в Redis изменения заметки"))
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}

// DeleteTaskById Обработка Delete-запроса типа /api/item/id для задач
// id в запросе передается в виде hex-строки, напр.:
// /api/tasks/item/id/?id=69959fd9aece410dd54f5739
func DeleteTaskById(ctx context.Context,
	taskRepo *mongodb.TaskRepository,
	client_redis *redis.Client,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var taskId RemindableId
		if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
			objectID, err := primitive.ObjectIDFromHex(taskId.Id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID задачи"})
				return
			}

			taskToBeDeleted, err := taskRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{"NotFound": fmt.Sprintf("Задачи с id=%s не существует.", taskId.Id)},
				)
				return
			}

			err = taskRepo.DeleteById(ctx, objectID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Удалена задача": taskToBeDeleted,
				})
				err = client_redis.SetJSON(
					ctx,
					fmt.Sprintf("Удалена задача %v", taskToBeDeleted.Name),
					taskToBeDeleted,
					time.Until(taskToBeDeleted.DueDate),
				)
				if err != nil {
					panic(fmt.Errorf("Ошибка логирования в Redis удаления задачи"))
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}

// DeleteNoteById Обработка Delete-запроса типа /api/item/id для заметок
// id в запросе передается в виде hex-строки, напр.:
// /api/notes/item/id/?id=69959fd9aece410dd54f5739
func DeleteNoteById(ctx context.Context,
	noteRepo *mongodb.NoteRepository,
	client_redis *redis.Client,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var noteId RemindableId
		if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
			objectID, err := primitive.ObjectIDFromHex(noteId.Id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заметки"})
				return
			}

			noteToBeDeleted, err := noteRepo.GetById(ctx, objectID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Заметки с id=%s не существует.", noteId.Id)})
				return
			}

			err = noteRepo.DeleteById(ctx, objectID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Удалена заметка": noteToBeDeleted,
				})
				err = client_redis.SetJSON(
					ctx,
					fmt.Sprintf("Удалена заметка %v", noteToBeDeleted.Name),
					noteToBeDeleted,
					time.Until(noteToBeDeleted.AlarmTimeStamp),
				)
				if err != nil {
					panic(fmt.Errorf("Ошибка логирования в Redis удаления заметки"))
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}
