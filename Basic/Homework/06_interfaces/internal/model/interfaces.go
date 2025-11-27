package model

type Remindable interface {
	String() string
	ChangeAlarm(string)
}
