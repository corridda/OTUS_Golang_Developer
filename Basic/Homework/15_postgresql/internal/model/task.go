package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"
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
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	InitTimeStamp time.Time `json:"initTimeStamp"`
	DueDate       time.Time `json:"dueDate"`
	Status        string    `json:"status"`
}

// NewTask генерирует и возвращает новую задачу
func NewTask(rows_count sql.NullInt64, name, descr, dueDate string) (Task, error) {
	userDueDate, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		log.Println("Введенная дата исполнения имеет не корректный формат.")
		if rows_count.Valid {
			return Task{
				Id:            int(rows_count.Int64 + 1),
				Name:          name,
				Description:   descr,
				InitTimeStamp: time.Now(),
				Status:        Created,
			}, nil
		} else {
			return Task{
				Id:            1,
				Name:          name,
				Description:   descr,
				InitTimeStamp: time.Now(),
				Status:        Created,
			}, nil
		}

	}
	if rows_count.Valid {
		return Task{
			Id:            int(rows_count.Int64 + 1),
			Name:          name,
			Description:   descr,
			InitTimeStamp: time.Now(),
			DueDate:       userDueDate,
			Status:        Created,
		}, nil
	}
	return Task{
		Id:            1,
		Name:          name,
		Description:   descr,
		InitTimeStamp: time.Now(),
		DueDate:       userDueDate,
		Status:        Created,
	}, nil
}

// String реализует repository.Remindable
func (myTask Task) String() string {
	return fmt.Sprintf(
		"Имя задачи: %v\nОписание задачи: %v\nДата постановки задачи: %v\nДата исполнения: %v\nСтатус: %v\n",
		myTask.Name, myTask.Description, myTask.InitTimeStamp.Format("02.01.2006"), myTask.DueDate.Format("02.01.2006"), myTask.Status,
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
