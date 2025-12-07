package main

import (
	"sync"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/07_channels_goroutines/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/07_channels_goroutines/internal/service"
)

// Искусственное заполенение списков задач и заметок
func createList(n int, chRemindable chan repository.Remindable, wg *sync.WaitGroup, mutex *sync.RWMutex) {
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			nextDay := time.Now().Add(time.Duration(time.Hour) * 24 * time.Duration(i*n))
			if i%2 == 0 {
				dueDate := nextDay.Format("02.01.2006")
				service.CreateNewRemindable("taskName", "taskDescr", dueDate, true, chRemindable, mutex)
			} else {
				alarmTimeStamp := nextDay.Format("02.01.2006 15:04")
				service.CreateNewRemindable("noteName", "noteDescr", alarmTimeStamp, false, chRemindable, mutex)
			}
		}(i)
		time.Sleep(time.Millisecond * 200) // для того, чтобы логгер успевал отрабатывать
	}
	wg.Wait()
}

func main() {
	wg := sync.WaitGroup{}
	mutex := sync.RWMutex{}
	stopLogger := make(chan struct{})

	ticker := time.NewTicker(time.Millisecond * 200)
	go service.LogRemidables(ticker, stopLogger, &mutex)

	n := 4
	chRemindable := make(chan repository.Remindable, n)
	createList(n, chRemindable, &wg, &mutex)
	wg.Wait()

	n = 2
	chRemindable = make(chan repository.Remindable, n)
	createList(n, chRemindable, &wg, &mutex)
	wg.Wait()

	stopLogger <- struct{}{}
	close(chRemindable)

	// service.PrintRemidables(&mutex)
}
