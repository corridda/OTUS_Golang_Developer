package service

import (
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/08_signals/internal/model"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/08_signals/internal/repository"
)

func CreateNewRemindable(
	name,
	descr,
	futurePoint string,
	isTask bool,
	chTask chan *model.Task,
	chNote chan *model.Note,
	mutex *sync.RWMutex,
) {
	chStop := make(chan any)
	if isTask {
		task := model.NewTask(name, descr, futurePoint)
		chTask <- &task
	} else {
		note := model.NewNote(name, descr, futurePoint)
		chNote <- &note
	}

	go repository.SaveRemindable(chTask, chNote, chStop, mutex)
	<-chStop
}
