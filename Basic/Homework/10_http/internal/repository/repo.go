package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Remindable interface {
	String() string
	ChangeAlarm(string)
}

type RemindableId struct {
	Id int `form:"id" binding:"required,gt=0"`
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

var Tasks = []model.Task{}
var Notes = []model.Note{}

func AppendJSONToFile(filename string, newRemindable Remindable, Id int) error {
	// прочитать существующие данные из файла
	fileBytes, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var tasks []model.Task
	var notes []model.Note
	var updatedBytes []byte

	switch value := newRemindable.(type) {
	case *model.Task:
		// десериализовать существующие данные из файла в слайс
		if len(fileBytes) > 0 {
			err = json.Unmarshal(fileBytes, &tasks)
			if err != nil {
				return fmt.Errorf("ошибка десериализации файла 'tasks.json': %w", err)
			}
		}
		// отсортировать по Id
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Id < tasks[j].Id
		})
		// добавать новую задачу в слайс
		v := *value
		v.Id = Id
		tasks = append(tasks, v)
		// сериализовать обновленный слайс обратно в JSON
		updatedBytes, err = json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			return fmt.Errorf("ошибка сериализации в JSON: %w", err)
		}
	case *model.Note:
		// десериализовать существующие данные из файла в слайс
		if len(fileBytes) > 0 {
			err = json.Unmarshal(fileBytes, &notes)
			if err != nil {
				return fmt.Errorf("ошибка десериализации файла 'notes.json': %w", err)
			}
		}
		// отсортировать по Id
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Id < notes[j].Id
		})
		// добавать новую заметку в слайс
		v := *value
		v.Id = Id
		notes = append(notes, v)
		// сериализовать обновленный слайс обратно в JSON
		updatedBytes, err = json.MarshalIndent(notes, "", "\t")
		if err != nil {
			return fmt.Errorf("ошибка сериализации в JSON: %w", err)
		}
	}
	// записать обновленные данные из слайса обратно в файл
	err = os.WriteFile(filename, updatedBytes, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл %s: %w", filename, err)
	}
	return nil
}

func writeToJSONFile(filename string, jsonData []byte) error {
	// записать обновленные данные из слайса обратно в файл
	err := os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл %s: %w", filename, err)
	}
	return nil
}

func UpdateJSONFile(filename string) error {
	var updatedBytes []byte
	var err error
	switch filename {
	case "tasks.json":
		// сериализовать обновленный слайс в JSON
		updatedBytes, err = json.MarshalIndent(Tasks, "", "\t")
		if err != nil {
			return fmt.Errorf("ошибка сериализации в JSON: %w", err)
		}

	case "notes.json":
		// сериализовать обновленный слайс в JSON
		updatedBytes, err = json.MarshalIndent(Notes, "", "\t")
		if err != nil {
			return fmt.Errorf("ошибка сериализации в JSON: %w", err)
		}
	}
	if err = writeToJSONFile(filename, updatedBytes); err != nil {
		return err
	}

	return nil
}

// создать объект типа, реализующего Remindable
func CreateNewRemindable(
	name,
	descr,
	futurePoint string,
	isTask bool,
) {
	chStop := make(chan any)
	var remindable Remindable
	if isTask {
		task := model.NewTask(name, descr, futurePoint)
		remindable = &task
	} else {
		note := model.NewNote(name, descr, futurePoint)
		remindable = &note
	}
	go SaveRemindable(remindable, chStop, &sync.RWMutex{})
	<-chStop
}

// Сохранить объект типа, реализующего Remindable в сооттв. срезе и json-файле
func SaveRemindable(
	remindable Remindable,
	chStop chan any,
	mutex *sync.RWMutex,
) {
	defer close(chStop)
	defer mutex.Unlock()

	mutex.Lock()
	r := remindable
	switch value := r.(type) {
	case *model.Task:
		v := *value
		if len(Tasks) != 0 {
			sort.Slice(Tasks, func(i, j int) bool {
				return Tasks[i].Id < Tasks[j].Id
			})
			v.Id = Tasks[len(Tasks)-1].Id + 1
		} else {
			v.Id = 1
		}
		Tasks = append(Tasks, v)
		if err := AppendJSONToFile("tasks.json", r, v.Id); err != nil {
			panic(fmt.Sprintf("ошибка при добавлении новой записи в файл %s: %v\n", "tasks.json", err))
		}
	case *model.Note:
		v := *value
		if len(Notes) != 0 {
			sort.Slice(Notes, func(i, j int) bool {
				return Notes[i].Id < Notes[j].Id
			})
			v.Id = Notes[len(Notes)-1].Id + 1
		} else {
			v.Id = 1
		}
		Notes = append(Notes, v)
		if err := AppendJSONToFile("notes.json", r, v.Id); err != nil {
			panic(fmt.Sprintf("ошибка при добавлении новой записи в файл %s: %v\n", "notes.json", err))
		}
	}
}

func FillTasksFromJSON(wg *sync.WaitGroup, errors chan<- error) {
	defer wg.Done()
	fileBytes, err := os.ReadFile("tasks.json")
	if err != nil && !os.IsNotExist(err) {
		errors <- fmt.Errorf("ошибка чтения файла: %w", err)
	}
	if len(fileBytes) > 0 {
		// отсортировать по Id
		defer func() {
			sort.Slice(Tasks, func(i, j int) bool {
				return Tasks[i].Id < Tasks[j].Id
			})
		}()
		err = json.Unmarshal(fileBytes, &Tasks)
		if err != nil {
			errors <- fmt.Errorf("ошибка десериализации файла 'tasks.json': %w", err)
		}
	}
}

func FillNotesFromJSON(wg *sync.WaitGroup, errors chan<- error) {
	defer wg.Done()
	fileBytes, err := os.ReadFile("notes.json")
	if err != nil && !os.IsNotExist(err) {
		errors <- fmt.Errorf("ошибка чтения файла: %w", err)
	}
	if len(fileBytes) > 0 {
		// отсортировать по Id
		defer func() {
			sort.Slice(Notes, func(i, j int) bool {
				return Notes[i].Id < Notes[j].Id
			})
		}()
		err = json.Unmarshal(fileBytes, &Notes)
		if err != nil {
			errors <- fmt.Errorf("ошибка десериализации файла 'notes.json': %w", err)
		}
	}
}

// Обработка Get-запроса типа /api/items для задач
func GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, Tasks)
}

// Обработка Get-запрос типа /api/item/id для задач
func GetTasksById(c *gin.Context) {
	var taskId RemindableId
	if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
		idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
			return task.Id == taskId.Id
		})
		if idx >= 0 {
			c.JSON(http.StatusOK, Tasks[idx])
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Обработка Get-запроса типа /api/items для заметок
func GetNotes(c *gin.Context) {
	c.JSON(http.StatusOK, Notes)
}

// Обработка Get-запроса типа /api/item/id для заметок
func GetNotesById(c *gin.Context) {
	var NoteId RemindableId
	if err := c.ShouldBindWith(&NoteId, binding.Query); err == nil {
		idx := slices.IndexFunc(Notes, func(note model.Note) bool {
			return note.Id == NoteId.Id
		})
		if idx >= 0 {
			c.JSON(http.StatusOK, Notes[idx])
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", NoteId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Обработка Post-запроса типа /api/item для задач
func PostNewTask(c *gin.Context) {
	newTask := NewTask{}
	err := c.ShouldBindJSON(&newTask)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	CreateNewRemindable(
		newTask.Name,
		newTask.Description,
		newTask.DueDate,
		true,
	)

	idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
		return task.Name == newTask.Name && task.Description == newTask.Description
	})

	if idx >= 0 {
		c.JSON(http.StatusOK, gin.H{
			"Создана новая задача": Tasks[idx],
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка создания новой задачи"})
	}
}

// Обработка Post-запроса типа /api/item для заметок
func PostNewNote(c *gin.Context) {
	newNote := NewNote{}
	err := c.ShouldBindJSON(&newNote)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	CreateNewRemindable(
		newNote.Name,
		newNote.Description,
		newNote.AlarmTimeStamp,
		false,
	)

	idx := slices.IndexFunc(Notes, func(note model.Note) bool {
		return note.Name == newNote.Name && note.Description == newNote.Description
	})

	if idx >= 0 {
		c.JSON(http.StatusOK, gin.H{
			"Создана новая заметка": Notes[idx],
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "Ошибка создания новой заметки"})
	}
}

// Обработка Put-запроса типа /api/item/id для задач
func PutTaskById(c *gin.Context) {
	var taskId RemindableId
	if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
		idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
			return task.Id == taskId.Id
		})
		if idx >= 0 {
			changingTask := ChangingTask{}
			err := c.ShouldBindJSON(&changingTask)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if changingTask.Name != "" && changingTask.Name != Tasks[idx].Name {
				Tasks[idx].Name = changingTask.Name
			}
			if changingTask.Description != "" && changingTask.Description != Tasks[idx].Description {
				Tasks[idx].Description = changingTask.Description
			}
			userDueDate, err := time.Parse("02.01.2006", changingTask.DueDate)
			if changingTask.DueDate != "" && userDueDate != Tasks[idx].DueDate {
				Tasks[idx].DueDate = userDueDate
			}
			Tasks[idx].Status = model.Updated
			UpdateJSONFile("tasks.json")
			c.JSON(http.StatusOK, gin.H{
				"Изменена задача": Tasks[idx],
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Обработка Put-запроса типа /api/item/id для заметок
func PutNoteById(c *gin.Context) {
	var noteId RemindableId
	if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
		idx := slices.IndexFunc(Notes, func(note model.Note) bool {
			return note.Id == noteId.Id
		})
		if idx >= 0 {
			changingNote := ChangingNote{}
			err := c.ShouldBindJSON(&changingNote)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if changingNote.Name != "" && changingNote.Name != Notes[idx].Name {
				Notes[idx].Name = changingNote.Name
			}
			if changingNote.Description != "" && changingNote.Description != Notes[idx].Description {
				Notes[idx].Description = changingNote.Description
			}
			userDueDate, err := time.Parse("02.01.2006 15:04", changingNote.AlarmTimeStamp)
			if changingNote.AlarmTimeStamp != "" && userDueDate != Notes[idx].AlarmTimeStamp {
				Notes[idx].AlarmTimeStamp = userDueDate
			}
			UpdateJSONFile("notes.json")
			c.JSON(http.StatusOK, gin.H{
				"Изменена заметка": Notes[idx],
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", noteId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Обработка Delete-запроса типа /api/item/id для задач
func DeleteTaskById(c *gin.Context) {
	var taskId RemindableId
	if err := c.ShouldBindWith(&taskId, binding.Query); err == nil {
		idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
			return task.Id == taskId.Id
		})
		if idx >= 0 {
			deletedTask := Tasks[idx]
			Tasks = slices.Delete(Tasks, idx, idx+1)
			UpdateJSONFile("tasks.json")
			c.JSON(http.StatusOK, gin.H{
				"Удалена задача": deletedTask,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Задачи с id=%d не существует.", taskId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Обработка Delete-запроса типа /api/item/id для заметок
func DeleteNoteById(c *gin.Context) {
	var noteId RemindableId
	if err := c.ShouldBindWith(&noteId, binding.Query); err == nil {
		idx := slices.IndexFunc(Notes, func(note model.Note) bool {
			return note.Id == noteId.Id
		})
		if idx >= 0 {
			deletedNote := Notes[idx]
			Notes = slices.Delete(Notes, idx, idx+1)
			UpdateJSONFile("notes.json")
			c.JSON(http.StatusOK, gin.H{
				"Удалена заметка": deletedNote,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"NotFound": fmt.Sprintf("Заметки с id=%d не существует.", noteId.Id)})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
