package service

import (
	"fmt"
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/09_files/internal/repository"
)

func PrintRemidables(mutex *sync.RWMutex) {
	mutex.RLock()
	fmt.Println("Имеющиеся задачи:")
	for _, t := range repository.Tasks {
		fmt.Println(t.String())
	}
	mutex.RUnlock()

	mutex.RLock()
	fmt.Println("\nИмеющиеся заметки:")
	for _, n := range repository.Notes {
		fmt.Println(n.String())
	}
	mutex.RUnlock()
}
