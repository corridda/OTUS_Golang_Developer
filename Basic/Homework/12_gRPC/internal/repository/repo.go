package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/12_gRPC/internal/model"
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
	var lastId int
	if isTask {
		if len(Tasks) == 0 {
			lastId = 0
		} else {
			lastId = Tasks[len(Tasks)-1].Id
		}
		task := model.NewTask(lastId, name, descr, futurePoint)
		remindable = &task
	} else {
		if len(Notes) == 0 {
			lastId = 0
		} else {
			lastId = Notes[len(Notes)-1].Id
		}
		note := model.NewNote(lastId, name, descr, futurePoint)
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

// Обработка Get-запроса для задач
func GetTasks() *[]model.Task {
	return &Tasks
}

// Обработка Get-запроса типа для заметок
func GetNotes() *[]model.Note {
	return &Notes
}

// Обработка Post-запроса для задач
func PostNewTask(name, description string, dueDate time.Time) *model.Task {
	newTask := NewTask{
		Name:        name,
		Description: description,
		DueDate:     dueDate.Format("02.01.2006"),
	}

	CreateNewRemindable(
		newTask.Name,
		newTask.Description,
		newTask.DueDate,
		true,
	)

	if Tasks[len(Tasks)-1].Name == newTask.Name && Tasks[len(Tasks)-1].Description == newTask.Description {
		return &Tasks[len(Tasks)-1]
	}
	return &model.Task{}
}

// Обработка Post-запроса для заметок
func PostNewNote(name, description string, alarmTimeStamp time.Time) *model.Note {
	newNote := NewNote{
		Name:           name,
		Description:    description,
		AlarmTimeStamp: alarmTimeStamp.Format("02.01.2006 15:04"),
	}

	CreateNewRemindable(
		newNote.Name,
		newNote.Description,
		newNote.AlarmTimeStamp,
		false,
	)

	if Notes[len(Notes)-1].Name == newNote.Name && Notes[len(Notes)-1].Description == newNote.Description {
		return &Notes[len(Notes)-1]
	}
	return &model.Note{}
}

// Обработка Put-запроса для задач
func PutTaskById(id int32, name, description string, dueDate time.Time) *model.Task {
	idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
		return task.Id == int(id)
	})
	if idx == -1 {
		return &model.Task{}
	}
	Tasks[idx].Name = name
	Tasks[idx].Description = description
	Tasks[idx].DueDate = dueDate
	err := UpdateJSONFile("tasks.json")
	if err != nil {
		panic("ошибка обновления файла tasks.json")
	}
	return &Tasks[idx]
}

// Обработка Put-запроса для заметок
func PutNoteById(id int32, name, description string, alarmTimeStamp time.Time) *model.Note {
	idx := slices.IndexFunc(Notes, func(note model.Note) bool {
		return note.Id == int(id)
	})
	if idx == -1 {
		return &model.Note{}
	}
	Notes[idx].Name = name
	Notes[idx].Description = description
	Notes[idx].AlarmTimeStamp = alarmTimeStamp
	err := UpdateJSONFile("notes.json")
	if err != nil {
		panic("ошибка обновления файла notes.json")
	}
	return &Notes[idx]
}

// Обработка Delete-запроса для задач
func DeleteTaskById(id int32) *model.Task {
	idx := slices.IndexFunc(Tasks, func(task model.Task) bool {
		return task.Id == int(id)
	})
	if idx == -1 {
		return &model.Task{}
	}
	deletedTask := model.Task{
		Id:            Tasks[idx].Id,
		Name:          Tasks[idx].Name,
		Description:   Tasks[idx].Description,
		InitTimeStamp: Tasks[idx].InitTimeStamp,
		DueDate:       Tasks[idx].DueDate,
		Status:        Tasks[idx].Status,
	}
	Tasks = slices.Delete(Tasks, idx, idx+1)
	err := UpdateJSONFile("tasks.json")
	if err != nil {
		panic("ошибка обновления файла tasks.json")
	}
	return &deletedTask
}

// Обработка Delete-запроса для заметок
func DeleteNoteById(id int32) *model.Note {
	idx := slices.IndexFunc(Notes, func(note model.Note) bool {
		return note.Id == int(id)
	})
	if idx == -1 {
		return &model.Note{}
	}
	deletedNote := model.Note{
		Id:             Notes[idx].Id,
		Name:           Notes[idx].Name,
		Description:    Notes[idx].Description,
		AlarmTimeStamp: Notes[idx].AlarmTimeStamp,
	}
	Notes = slices.Delete(Notes, idx, idx+1)
	err := UpdateJSONFile("notes.json")
	if err != nil {
		panic("ошибка обновления файла notes.json")
	}
	return &deletedNote
}
