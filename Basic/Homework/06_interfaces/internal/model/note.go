package model

import (
	"fmt"
	"time"
)

type Note struct {
	name           string
	description    string
	alarmTimeStamp time.Time // Сигнал напоминания в эту дату-время
}

func NewNote(name, descr, alarmDateTime string) Note {
	userDueDate, err := time.Parse("02.01.2006 15:04", alarmDateTime)
	if err != nil {
		fmt.Println("Введенные дата и время напоминания имеют не корректный формат.")
		return Note{
			name:        name,
			description: descr,
		}
	} else {
		return Note{
			name:           name,
			description:    descr,
			alarmTimeStamp: userDueDate,
		}
	}
}

func (myNote Note) String() string {
	return fmt.Sprintf(
		"Имя заметки: %v\nОписание заметки: %v\nДата и время напоминания: %v\n",
		myNote.name, myNote.description, myNote.alarmTimeStamp.Format("02.01.2006 15:04"),
	)
}

// ChangeAlarm implements Remindable.
func (myNote *Note) ChangeAlarm(new_date_time string) {
	userDateTime, err := time.Parse("02.01.2006 15:04", new_date_time)
	if err != nil {
		fmt.Println("Введенные дата и время напоминания имеют не корректный формат.\nВведите требуемые значения даты и времени в соответствии с форматом: ДД-ММ-ГГГГ ЧЧ:ММ.")
	} else {
		myNote.alarmTimeStamp = userDateTime
	}
}
