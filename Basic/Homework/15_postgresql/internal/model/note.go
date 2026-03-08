package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Note struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	AlarmTimeStamp time.Time `json:"alarmTimeStamp"` // Сигнал напоминания в эту дату-время
}

func NewNote(ctx context.Context, db *sql.DB, name, descr, alarmDateTime string) (Note, error) {
	var rows_count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) from notes").Scan(&rows_count)
	if err != nil {
		return Note{}, fmt.Errorf("Ошибка считывания количества строк из БД: %v", err)
	}

	userDueDate, err := time.Parse("02.01.2006 15:04", alarmDateTime)
	if err != nil {
		log.Println("Введенные дата и время напоминания имеют не корректный формат.")
		return Note{
			Id:          rows_count + 1,
			Name:        name,
			Description: descr,
		}, nil
	} else {
		return Note{
			Id:             rows_count + 1,
			Name:           name,
			Description:    descr,
			AlarmTimeStamp: userDueDate,
		}, nil
	}
}

// String реализует repository.Remindable
func (myNote Note) String() string {
	return fmt.Sprintf(
		"Имя заметки: %v\nОписание заметки: %v\nДата и время напоминания: %v\n",
		myNote.Name, myNote.Description, myNote.AlarmTimeStamp.Format("02.01.2006 15:04"),
	)
}

// ChangeAlarm реализует repository.Remindable
func (myNote *Note) ChangeAlarm(new_date_time string) {
	userDateTime, err := time.Parse("02.01.2006 15:04", new_date_time)
	if err != nil {
		fmt.Println("Введенные дата и время напоминания имеют не корректный формат.\nВведите требуемые значения даты и времени в соответствии с форматом: ДД-ММ-ГГГГ ЧЧ:ММ.")
	} else {
		myNote.AlarmTimeStamp = userDateTime
	}
}
