package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	Id             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
	AlarmTimeStamp time.Time          `bson:"alarmTimeStamp" json:"alarmTimeStamp"` // Сигнал напоминания в эту дату-время
}

func NewNote(name, descr, alarmDateTime string) Note {
	userDueDate, err := time.Parse("02.01.2006 15:04", alarmDateTime)
	if err != nil {
		fmt.Println("Введенные дата и время напоминания имеют не корректный формат.")
		return Note{
			Name:        name,
			Description: descr,
		}
	} else {
		return Note{
			Name:           name,
			Description:    descr,
			AlarmTimeStamp: userDueDate,
		}
	}
}

// String реализует repository.Remindable
func (myNote Note) String() string {
	return fmt.Sprintf(
		"Id заметки: %v\nИмя заметки: %v\nОписание заметки: %v\nДата и время напоминания: %v\n",
		myNote.Id, myNote.Name, myNote.Description, myNote.AlarmTimeStamp.Format("02.01.2006 15:04"),
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
