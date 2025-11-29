package main

import (
	"fmt"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/06_interfaces/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/06_interfaces/internal/service"
)

func main() {
	for i := 1; i < 6; i++ {
		nextDay := time.Now().Add(time.Duration(time.Hour) * 24 * time.Duration(i))
		if i%2 == 0 {
			tasksNum := len(repository.Tasks)
			dueDate := nextDay.Format("02.01.2006")
			service.CreateNewRemindable(fmt.Sprintf("task №%d", tasksNum+1), "taskDescr", dueDate, true)
		} else {
			notesNum := len(repository.Notes)
			alarmTimeStamp := nextDay.Format("02.01.2006 15:04")
			service.CreateNewRemindable(fmt.Sprintf("note №%d", notesNum+1), "noteDescr", alarmTimeStamp, false)
		}
		time.Sleep(100 * time.Millisecond)
	}

	repository.PrintRemidables()
}
