package model

import (
	"context"
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

func NewTask(ctx context.Context, db *sql.DB, name, descr, dueDate string) (Task, error) {
	var rows_count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) from tasks").Scan(&rows_count)
	if err != nil {
		return Task{}, fmt.Errorf("Ошибка считывания количества строк из БД: %v", err)
	}

	userDueDate, err := time.Parse("02.01.2006", dueDate)
	if err != nil {
		log.Println("Введенная дата исполнения имеет не корректный формат.")
		return Task{
			Id:            rows_count + 1,
			Name:          name,
			Description:   descr,
			InitTimeStamp: time.Now(),
			Status:        Created,
		}, nil
	} else {
		return Task{
			Id:            rows_count + 1,
			Name:          name,
			Description:   descr,
			InitTimeStamp: time.Now(),
			DueDate:       userDueDate,
			Status:        Created,
		}, nil
	}
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
