package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/service"
)

// Искусственное заполенение списков задач и заметок
func createList(
	n int,
	ctx context.Context,
	chRemindable chan repository.Remindable,
	mutex *sync.RWMutex,
) {
	wg := sync.WaitGroup{}
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				nextDay := time.Now().Add(time.Duration(time.Hour) * 24 * time.Duration(i*n))
				if i%2 == 0 {
					dueDate := nextDay.Format("02.01.2006")
					length := len(repository.Tasks)
					service.CreateNewRemindable(
						fmt.Sprintf("taskName %d", length+1),
						"taskDescr",
						dueDate,
						true,
						chRemindable,
						mutex,
					)
				} else {
					alarmTimeStamp := nextDay.Format("02.01.2006 15:04")
					length := len(repository.Notes)
					service.CreateNewRemindable(
						fmt.Sprintf("noteName %d", length+1),
						"noteDescr",
						alarmTimeStamp,
						false,
						chRemindable,
						mutex,
					)
				}
			}
		}(i)
		time.Sleep(time.Millisecond * 200) // для того, чтобы логгер успевал отрабатывать
	}
	wg.Wait()
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

func main() {
	// создание json-файлов для хранения задач и заметок
	createFiles("tasks.json", "notes.json")

	wg := sync.WaitGroup{}
	mutex := sync.RWMutex{}
	errsTasks := make(chan error, 1)
	errsNotes := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

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
	go service.LogRemidables(ctx, ticker, &mutex)

	// генерация новых задач и заметок
	n := 10
	chRemindable := make(chan repository.Remindable, n)
	go createList(n, ctx, chRemindable, &mutex)

	sig := <-sigChan
	cancel()
	time.Sleep(time.Second)
	fmt.Printf("Программа завершает свою работу по сигналу %v\n", sig)

	close(chRemindable)

	// repository.PrintRemidables(&mutex)
}
