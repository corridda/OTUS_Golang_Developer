package repository

import (
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/07_channels_goroutines/internal/model"
)

type Remindable interface {
	String() string
	ChangeAlarm(string)
}

var Tasks = []model.Task{}
var Notes = []model.Note{}

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
	case *model.Note:
		Notes = append(Notes, *value)
	}
}
