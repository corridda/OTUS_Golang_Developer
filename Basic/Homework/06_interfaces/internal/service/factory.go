package service

import (
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/06_interfaces/internal/model"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/06_interfaces/internal/repository"
)

func CreateNewRemindable(name, descr, futurePoint string, isTask bool) {
	var remindable model.Remindable
	if isTask {
		task := model.NewTask(name, descr, futurePoint)
		remindable = &task
	} else {
		note := model.NewNote(name, descr, futurePoint)
		remindable = &note
	}
	repository.SaveRemindable(remindable)
}
