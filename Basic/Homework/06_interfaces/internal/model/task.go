package model

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

func NewTask(name, descr, dueDate string) Task {
	userDueDate, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		fmt.Println("Введенная дата исполнения имеет не корректный формат.")
		return Task{
			name:          name,
			description:   descr,
			initTimeStamp: time.Now(),
			status:        Created,
		}
	} else {
		return Task{
			name:          name,
			description:   descr,
			initTimeStamp: time.Now(),
			dueDate:       userDueDate,
			status:        Created,
		}
	}
}

func (myTask Task) String() string {
	return fmt.Sprintf(
		"Имя задачи: %v\nОписание задачи: %v\nДата постановки задачи: %v\nДата исполнения: %v\nСтатус: %v\n",
		myTask.name, myTask.description, myTask.initTimeStamp.Format("02.01.2006"), myTask.dueDate.Format("02.01.2006"), myTask.status,
	)
}

// ChangeAlarm implements Remindable.
func (myTask *Task) ChangeAlarm(new_date string) {
	userDate, err := time.Parse("02.01.2006", new_date)
	if err != nil {
		fmt.Println("Введенная дата исполнения имеет не корректный формат.\nВведите требуемое значение даты в соответствии с форматом: ДД-ММ-ГГГГ.")
	} else {
		myTask.dueDate = userDate
	}
}
