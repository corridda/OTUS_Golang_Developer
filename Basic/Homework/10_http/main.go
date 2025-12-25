package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/service"
)

func createFiles(fileNames ...string) {
	wg := sync.WaitGroup{}
	for _, fileName := range fileNames {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			// O_EXCL - used with O_CREATE, file must not exist.
			file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				// если файл существует
				if errors.Is(err, os.ErrExist) {
					return
				}
				// при всяких других ошибках
				panic(err)
			}
			fmt.Printf("Файл %s создан.\n", fileName)
			defer file.Close()
		}(fileName)
	}
	wg.Wait()
	fmt.Println()
}

func main() {
	// создание json-файлов для хранения задач и заметок
	createFiles("tasks.json", "notes.json")

	wg := sync.WaitGroup{}
	mutex := sync.RWMutex{}
	errsTasks := make(chan error, 1)
	errsNotes := make(chan error, 1)

	// заполнение срезов задач и заметок соответствующими данными из json-файлов
	wg.Add(2)
	go repository.FillTasksFromJSON(&wg, errsTasks)
	go repository.FillNotesFromJSON(&wg, errsNotes)

	go func() {
		wg.Wait()
		close(errsTasks)
		close(errsNotes)
	}()

	if err := <-errsTasks; err != nil {
		panic(fmt.Sprintf("ошибка наполнения repository.Tasks из файла tasks.json: %v", err))
	}
	if err := <-errsNotes; err != nil {
		panic(fmt.Sprintf("ошибка наполнения repository.Notes из файла notes.json: %v", err))
	}

	// запуск логгера для новых задач/заметок
	ticker := time.NewTicker(time.Millisecond * 200)
	go service.LogRemidables(ticker, &mutex)

	// Создаём роутер
	r := gin.Default()
	api := r.Group("/api")
	apiTasks := api.Group("/tasks")
	apiNotes := api.Group("/notes")
	apiTasks.GET("items", repository.GetTasks)            // /api/tasks/items
	apiTasks.GET("item/id", repository.GetTasksById)      // /api/tasks/item/id/?id=<id_integer_number>
	apiNotes.GET("items", repository.GetNotes)            // /api/notes/items
	apiNotes.GET("item/id", repository.GetNotesById)      // /api/notes/item/id/?id=<id_integer_number>
	apiTasks.POST("item", repository.PostNewTask)         // /api/tasks/item
	apiNotes.POST("item", repository.PostNewNote)         // /api/tasks/item
	apiTasks.PUT("item/id", repository.PutTaskById)       // /api/tasks/item/id/?id=<id_integer_number>
	apiNotes.PUT("item/id", repository.PutNoteById)       // /api/notes/item/id/?id=<id_integer_number>
	apiTasks.DELETE("item/id", repository.DeleteTaskById) // /api/tasks/item/id/?id=<id_integer_number>
	apiNotes.DELETE("item/id", repository.DeleteNoteById) // /api/notes/item/id/?id=<id_integer_number>

	// Запуск сервера на :8080
	r.Run(":8080")

}
