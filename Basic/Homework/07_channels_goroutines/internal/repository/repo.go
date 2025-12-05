package repository

import (
	"fmt"
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/07_channels_goroutines/internal/model"
)

var Tasks = []model.Task{}
var Notes = []model.Note{}

func SaveRemindable(
	chT <-chan *model.Task,
	chN <-chan *model.Note,
	chStop chan any,
	mutex *sync.RWMutex,
) {
	defer close(chStop)
	defer mutex.Unlock()

	mutex.Lock()
	select {
	case newTask := <-chT:
		Tasks = append(Tasks, *newTask)
	case newNote := <-chN:
		Notes = append(Notes, *newNote)
	}
}

func PrintRemidables(mutex *sync.RWMutex) {
	mutex.RLock()
	fmt.Println("Имеющиеся задачи:")
	for _, t := range Tasks {
		fmt.Println(t.String())
	}
	mutex.RUnlock()

	mutex.RLock()
	fmt.Println("\nИмеющиеся заметки:")
	for _, n := range Notes {
		fmt.Println(n.String())
	}
	mutex.RUnlock()
}
