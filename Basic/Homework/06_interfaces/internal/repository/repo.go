package repository

import (
	"fmt"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/06_interfaces/internal/model"
)

var Tasks = []model.Task{}
var Notes = []model.Note{}

func SaveRemindable(r model.Remindable) {
	switch value := r.(type) {
	case *model.Task:
		Tasks = append(Tasks, *value)
	case *model.Note:
		Notes = append(Notes, *value)
	}
}

func PrintRemidables() {
	fmt.Println("Имеющиеся задачи:")
	for _, t := range Tasks {
		fmt.Println(t.String())
	}
	fmt.Println("\nИмеющиеся заметки:")
	for _, n := range Notes {
		fmt.Println(n.String())
	}
}
