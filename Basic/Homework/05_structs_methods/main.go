package main

import (
	"fmt"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/05_structs_methods/internal/model/menu"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/05_structs_methods/internal/model/task"
)

func main() {
	myMenu := menu.NewMenu()
	fmt.Println(myMenu)

	dueDate := task.NewMyDate(31, 12, 2025)
	newTask := task.NewTask("taskName", "taskDescr", dueDate)
	fmt.Println((&newTask).String())

	// fmt.Println(time.Now())
	// fmt.Println(time.Now().Local())
	// fmt.Println(time.Now().Location())
}
