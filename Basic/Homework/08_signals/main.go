package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/08_signals/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/08_signals/internal/service"
)

// Искусственное заполенение списков задач и заметок
func createList(
	n int,
	ctx context.Context,
	chRemindable chan repository.Remindable,
	wg *sync.WaitGroup,
	mutex *sync.RWMutex,
) {
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
					service.CreateNewRemindable("taskName", "taskDescr", dueDate, true, chRemindable, mutex)
				} else {
					alarmTimeStamp := nextDay.Format("02.01.2006 15:04")
					service.CreateNewRemindable("noteName", "noteDescr", alarmTimeStamp, false, chRemindable, mutex)
				}
			}
		}(i)
		time.Sleep(time.Millisecond * 200) // для того, чтобы логгер успевал отрабатывать
	}
	wg.Wait()
}

func main() {
	wg := sync.WaitGroup{}
	mutex := sync.RWMutex{}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Millisecond * 200)
	go service.LogRemidables(ctx, ticker, &mutex)

	n := 20
	chRemindable := make(chan repository.Remindable, n)
	go createList(n, ctx, chRemindable, &wg, &mutex)

	sig := <-sigChan
	cancel()
	time.Sleep(time.Second)
	fmt.Printf("Программа завершает свою работу по сигналу %v\n", sig)

	close(chRemindable)

	// repository.PrintRemidables(&mutex)
}
