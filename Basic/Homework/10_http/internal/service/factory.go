package service

import (
	"sync"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/model"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/10_http/internal/repository"
)

func CreateNewRemindable(
	name,
	descr,
	futurePoint string,
	isTask bool,
	chRemindable chan repository.Remindable,
	mutex *sync.RWMutex,
) {
	chStop := make(chan any)
	var remindable repository.Remindable
	if isTask {
		task := model.NewTask(name, descr, futurePoint)
		remindable = &task
	} else {
		note := model.NewNote(name, descr, futurePoint)
		remindable = &note
	}
	chRemindable <- remindable

	go repository.SaveRemindable(chRemindable, chStop, mutex)
	<-chStop
}
