package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/12_gRPC/proto_api/pkg/grpc/v1/remindables_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func stringlifyTask(
	id int32,
	name string,
	description string,
	initTimeStamp string,
	dueDate string,
	status string,
) string {
	return fmt.Sprintf("\n{\n\tId: %v\n\tName: %v\n\tDescription: %v\n\tInitTimeStamp: %v\n\tDueDate: %v\n\tStatus: %v\n}\n", id, name, description, initTimeStamp, dueDate, status)
}

func stringlifyPostedTask(resp *remindables_api.PostNewTaskResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	initTimeStamp := resp.InitTimeStamp.AsTime().Format("02.01.2006 15:04")
	dueDate := resp.DueDate.AsTime().Format("02.01.2006")
	status := resp.Status
	return stringlifyTask(id, name, description, initTimeStamp, dueDate, status)
}

func stringlifyGottenTask(resp *remindables_api.GetTaskResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	initTimeStamp := resp.InitTimeStamp.AsTime().Format("02.01.2006 15:04")
	dueDate := resp.DueDate.AsTime().Format("02.01.2006")
	status := resp.Status
	return stringlifyTask(id, name, description, initTimeStamp, dueDate, status)
}

func stringlifyPutTask(resp *remindables_api.PutTaskResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	initTimeStamp := resp.InitTimeStamp.AsTime().Format("02.01.2006 15:04")
	dueDate := resp.DueDate.AsTime().Format("02.01.2006")
	status := resp.Status
	return stringlifyTask(id, name, description, initTimeStamp, dueDate, status)
}

func stringlifyDeletedTask(resp *remindables_api.DeleteTaskResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	initTimeStamp := resp.InitTimeStamp.AsTime().Format("02.01.2006 15:04")
	dueDate := resp.DueDate.AsTime().Format("02.01.2006")
	status := resp.Status
	return stringlifyTask(id, name, description, initTimeStamp, dueDate, status)
}

func stringlifyNote(
	id int32,
	name string,
	description string,
	alarmTimeStamp string,
) string {
	return fmt.Sprintf("\n{\n\tId: %v\n\tName: %v\n\tDescription: %v\n\tAlarmTimeStamp: %v\n}\n", id, name, description, alarmTimeStamp)
}

func stringlifyPostedNote(resp *remindables_api.PostNewNoteResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	alarmTimeStamp := resp.AlarmTimeStamp.AsTime().Format("02.01.2006 15:04")
	return stringlifyNote(id, name, description, alarmTimeStamp)
}

func stringlifyGottenNote(resp *remindables_api.GetNoteResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	alarmTimeStamp := resp.AlarmTimeStamp.AsTime().Format("02.01.2006 15:04")
	return stringlifyNote(id, name, description, alarmTimeStamp)
}

func stringlifyPutNote(resp *remindables_api.PutNoteResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	alarmTimeStamp := resp.AlarmTimeStamp.AsTime().Format("02.01.2006 15:04")
	return stringlifyNote(id, name, description, alarmTimeStamp)
}

func stringlifyDeletedNote(resp *remindables_api.DeleteNoteResponse) string {
	id := resp.Id
	name := resp.Name
	description := resp.Description
	alarmTimeStamp := resp.AlarmTimeStamp.AsTime().Format("02.01.2006 15:04")
	return stringlifyNote(id, name, description, alarmTimeStamp)
}

func main() {
	conn, err := grpc.NewClient("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	cl := remindables_api.NewRemindablesServiceClient(conn)

	// Создать новую задачу
	dueDate, _ := time.Parse("02.01.2006", "03.04.2026")
	res0, err := cl.PostNewTask(context.Background(), &remindables_api.PostNewTaskRequest{
		Name:        "taskName 1",
		Description: "taskDescr 1",
		DueDate:     timestamppb.New(dueDate),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно создана новая задача:%s", stringlifyPostedTask(res0))

	// Создать новую задачу
	dueDate, _ = time.Parse("02.01.2006", "04.04.2026")
	res1, err := cl.PostNewTask(context.Background(), &remindables_api.PostNewTaskRequest{
		Name:        "taskName 2",
		Description: "taskDescr 2",
		DueDate:     timestamppb.New(dueDate),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно создана новая задача:%s", stringlifyPostedTask(res1))

	// Создать новую заметку
	alarmTimeStamp, _ := time.Parse("02.01.2006 15:04", "03.04.2026 20:00")
	res2, err := cl.PostNewNote(context.Background(), &remindables_api.PostNewNoteRequest{
		Name:           "noteName 1",
		Description:    "noteDescr 1",
		AlarmTimeStamp: timestamppb.New(alarmTimeStamp),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно создана новая заметка:%s", stringlifyPostedNote(res2))

	// Создать новую заметку
	alarmTimeStamp, _ = time.Parse("02.01.2006 15:04", "04.04.2026 20:00")
	res3, err := cl.PostNewNote(context.Background(), &remindables_api.PostNewNoteRequest{
		Name:           "noteName 2",
		Description:    "noteDescr 2",
		AlarmTimeStamp: timestamppb.New(alarmTimeStamp),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно создана новая заметка:%s", stringlifyPostedNote(res3))

	// Считать все имеющиеся задачи
	res4, err := cl.GetTasks(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		task, err := res4.Recv()
		if err == io.EOF {
			break // Stream is finished
		}
		if err != nil {
			panic("ошибка получения задач из хранилища")
		}
		fmt.Printf("\nЗадача:%v\n", stringlifyGottenTask(task))
	}

	// Считать все имеющиеся заметки
	res5, err := cl.GetNotes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		note, err := res5.Recv()
		if err == io.EOF {
			break // Stream is finished
		}
		if err != nil {
			panic("ошибка получения заметок из хранилища")
		}
		fmt.Printf("\nЗаметкa:%v\n", stringlifyGottenNote(note))
	}

	// Считать задачу по ее ID
	res6, err := cl.GetTasksById(context.Background(), &remindables_api.GetTaskRequest{
		Id: 1,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно считана задача по ID=%d:%s", 1, stringlifyGottenTask(res6))

	// Считать заметку по ее ID
	res7, err := cl.GetNotesById(context.Background(), &remindables_api.GetNoteRequest{
		Id: 1,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно считана заметка по ID=%d:%s", 1, stringlifyGottenNote(res7))

	// Изменить задачу по ее ID
	dueDate, _ = time.Parse("02.01.2006", "31.12.2026")
	res8, err := cl.PutTaskById(context.Background(), &remindables_api.PutTaskRequest{
		Id:          1,
		Name:        "new taskName 1",
		Description: "new taskDescr 1",
		DueDate:     timestamppb.New(dueDate),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно изменена задача с ID=%d:%s", 1, stringlifyPutTask(res8))

	// Изменить заметку по ее ID
	alarmTimeStamp, _ = time.Parse("02.01.2006 15:04", "31.12.2026 23:59")
	res9, err := cl.PutNoteById(context.Background(), &remindables_api.PutNoteRequest{
		Id:             1,
		Name:           "new noteName 1",
		Description:    "new noteDescr 1",
		AlarmTimeStamp: timestamppb.New(alarmTimeStamp),
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно изменена заметка с ID=%d:%s", 1, stringlifyPutNote(res9))

	// Удалить задачу по ее ID
	res10, err := cl.DeleteTaskById(context.Background(), &remindables_api.DeleteTaskRequest{
		Id: 2,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно удалена задача с ID=%d:%s", 2, stringlifyDeletedTask(res10))

	// Удалить заметку по ее ID
	res11, err := cl.DeleteNoteById(context.Background(), &remindables_api.DeleteNoteRequest{
		Id: 2,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Успешно удалена заметка с ID=%d:%s", 2, stringlifyDeletedNote(res11))
}
