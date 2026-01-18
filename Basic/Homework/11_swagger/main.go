package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/corridda/OTUS_Golang_Developer/Basic/Homework/11_swagger/docs"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/11_swagger/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/11_swagger/internal/service"
)

// @title Программа для управления задачами
// @version 1
// @description API Server
// @host localhost:8080/

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

	// Инициализация обработчика запросов
	// Именно он будет отвечать за хендлеры и обрабатывать каждый запрос.
	handle := repository.New()

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Инициализация маршрутов и обработчиков запросов
	// Функция нужна чтобы добавить каждый хендлер в сервер
	repository.InitHandler(r, handle)

	// Запуск сервера на :8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}

}

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
