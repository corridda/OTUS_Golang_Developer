package task

import (
	"fmt"
	"time"
)

type MyDate struct {
	day   int
	month int
	year  int
}

func NewMyDate(day int, month int, year int) MyDate {
	return MyDate{
		day:   day,
		month: month,
		year:  year,
	}
}

type Task struct {
	name          string
	description   string
	initTimeStamp time.Time
	dueDate       time.Time
}

func (myTaskP *Task) SetDueDate(myDate MyDate) {
	(*myTaskP).dueDate = time.Date(
		myDate.year,
		time.Month(myDate.month),
		myDate.day,
		23,
		59,
		59,
		0,
		time.Now().Location(),
	)
}

func NewTask(name string, descr string, dueDate MyDate) Task {
	myTask := Task{
		name:          name,
		description:   descr,
		initTimeStamp: time.Now(),
	}
	(&myTask).SetDueDate(dueDate)
	return myTask
}

func (myTask *Task) String() string {
	return fmt.Sprintf(
		"\n\tИмя задачи: %v\n\tОписание задачи: %v\n\tДата постановки задачи: %v\n\tДата исполнения: %v",
		(*myTask).name, (*myTask).description, (*myTask).initTimeStamp.Format(time.DateTime), (*myTask).dueDate.Format(time.DateTime),
	)
}
