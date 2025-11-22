package task

import (
	"fmt"
	"time"
)

const (
	Created   = "Создана"
	Seen      = "Просмотрена"
	InProcess = "В работе"
	Suspended = "Приостановлена"
	Submitted = "Решена, ждет контроля"
	Completed = "Завершена"
	Cancelled = "Отменена"
	Returned  = "Возвращена на доработку"
	Backlog   = "Бэклог"
)

type Task struct {
	name          string
	description   string
	initTimeStamp time.Time
	dueDate       time.Time
	status        string
}

func NewTask(name, descr string, dueDate string) Task {
	userDueDate, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		panic("Введенная дата исполнения имеет не корректный формат.")
	}
	myTask := Task{
		name:          name,
		description:   descr,
		initTimeStamp: time.Now(),
		dueDate:       userDueDate,
		status:        Created,
	}
	return myTask
}

func (myTask *Task) String() string {
	return fmt.Sprintf(
		"Имя задачи: %v\nОписание задачи: %v\nДата постановки задачи: %v\nДата исполнения: %v\nСтатус: %v\n",
		myTask.name, myTask.description, myTask.initTimeStamp.Format(time.DateTime), myTask.dueDate.Format(time.DateTime), myTask.status,
	)
}
