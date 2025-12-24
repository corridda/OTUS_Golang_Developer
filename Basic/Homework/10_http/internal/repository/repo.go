package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/model"
)

type Remindable interface {
	String() string
	ChangeAlarm(string)
}

var Tasks = []model.Task{}
var Notes = []model.Note{}

func AppendJSONToFile(filename string, newRemindable Remindable) error {
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
		// добавать новую задачу в слайс
		tasks = append(tasks, *value)
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
		// добавать новую заметку в слайс
		notes = append(notes, *value)
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

func SaveRemindable(
	chRemindable chan Remindable,
	chStop chan any,
	mutex *sync.RWMutex,
) {
	defer close(chStop)
	defer mutex.Unlock()

	mutex.Lock()
	r := <-chRemindable
	switch value := r.(type) {
	case *model.Task:
		Tasks = append(Tasks, *value)
		if err := AppendJSONToFile("tasks.json", r); err != nil {
			panic(fmt.Sprintf("ошибка при добавлении новой записи в файл %s: %v\n", "tasks.json", err))
		}
	case *model.Note:
		Notes = append(Notes, *value)
		if err := AppendJSONToFile("notes.json", r); err != nil {
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
		err = json.Unmarshal(fileBytes, &Notes)
		if err != nil {
			errors <- fmt.Errorf("ошибка десериализации файла 'notes.json': %w", err)
		}
	}
}
