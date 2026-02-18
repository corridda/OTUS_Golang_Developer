package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Created   = "Создана"
	Updated   = "Изменена"
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
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	InitTimeStamp time.Time          `bson:"initTimeStamp" json:"initTimeStamp"`
	DueDate       time.Time          `bson:"dueDate" json:"dueDate"`
	Status        string             `bson:"status" json:"status"`
}

func NewTask(name, descr, dueDate string) Task {
	userDueDate, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		fmt.Println("Введенная дата исполнения имеет не корректный формат.")
		return Task{
			Name:          name,
			Description:   descr,
			InitTimeStamp: time.Now(),
			Status:        Created,
		}
	} else {
		return Task{
			Name:          name,
			Description:   descr,
			InitTimeStamp: time.Now(),
			DueDate:       userDueDate,
			Status:        Created,
		}
	}
}

// String реализует repository.Remindable
func (myTask Task) String() string {
	return fmt.Sprintf(
		"Id задачи: %v\nИмя задачи: %v\nОписание задачи: %v\nДата постановки задачи: %v\nДата исполнения: %v\nСтатус: %v\n",
		myTask.Id, myTask.Name, myTask.Description, myTask.InitTimeStamp.Format("02.01.2006"), myTask.DueDate.Format("02.01.2006"), myTask.Status,
	)
}

// ChangeAlarm реализует repository.Remindable
func (myTask *Task) ChangeAlarm(new_date string) {
	userDate, err := time.Parse("02.01.2006", new_date)
	if err != nil {
		fmt.Println("Введенная дата исполнения имеет не корректный формат.\nВведите требуемое значение даты в соответствии с форматом: ДД-ММ-ГГГГ.")
	} else {
		myTask.DueDate = userDate
	}
}
