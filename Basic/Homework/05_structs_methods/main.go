package main

import (
	"fmt"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/05_structs_methods/internal/model/task"
)

var menu = struct {
	add    string
	view   string
	update string
	delete string
}{
	add:    "Добавить",
	view:   "Отобразить",
	update: "Обновить",
	delete: "Удалить",
}

func main() {
	fmt.Printf("Меню: %v\n\n", menu)

	userInputDueDate := "01.01.2026"
	newTask := task.NewTask("taskName", "taskDescr", userInputDueDate)
	fmt.Println((&newTask).String())
}
