package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/12_gRPC/internal/model"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/12_gRPC/internal/repository"
	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/12_gRPC/proto_api/pkg/grpc/v1/remindables_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	remindables_api.UnimplementedRemindablesServiceServer
}

// GetTasks implements remindables_api.RemindablesServiceClient.
func (s *server) GetTasks(
	args *emptypb.Empty,
	stream grpc.ServerStreamingServer[remindables_api.GetTaskResponse],
) error {
	tasks := repository.GetTasks()
	for _, task := range *tasks {
		stream.Send(&remindables_api.GetTaskResponse{
			Id:            int32(task.Id),
			Name:          task.Name,
			Description:   task.Description,
			InitTimeStamp: timestamppb.New(task.InitTimeStamp),
			DueDate:       timestamppb.New(task.DueDate),
			Status:        task.Status,
		})
	}

	return nil
}

// GetNotes implements remindables_api.RemindablesServiceClient.
func (s *server) GetNotes(
	args *emptypb.Empty,
	stream grpc.ServerStreamingServer[remindables_api.GetNoteResponse],
) error {
	notes := repository.GetNotes()
	for _, note := range *notes {
		stream.Send(&remindables_api.GetNoteResponse{
			Id:             int32(note.Id),
			Name:           note.Name,
			Description:    note.Description,
			AlarmTimeStamp: timestamppb.New(note.AlarmTimeStamp),
		})
	}

	return nil
}

// GetTasksById implements remindables_api.RemindablesServiceClient.
func (s *server) GetTasksById(
	ctx context.Context,
	userRequest *remindables_api.GetTaskRequest,
) (*remindables_api.GetTaskResponse, error) {
	tasks := *repository.GetTasks()
	taskID := userRequest.GetId()
	idx := slices.IndexFunc(tasks, func(task model.Task) bool {
		return task.Id == int(taskID)
	})
	return &remindables_api.GetTaskResponse{
		Id:            int32(tasks[idx].Id),
		Name:          tasks[idx].Name,
		Description:   tasks[idx].Description,
		InitTimeStamp: timestamppb.New(tasks[idx].InitTimeStamp),
		DueDate:       timestamppb.New(tasks[idx].DueDate),
		Status:        tasks[idx].Status,
	}, nil
}

// GetNotesById implements remindables_api.RemindablesServiceClient.
func (s *server) GetNotesById(
	ctx context.Context,
	userRequest *remindables_api.GetNoteRequest,
) (*remindables_api.GetNoteResponse, error) {
	notes := *repository.GetNotes()
	noteID := userRequest.GetId()
	idx := slices.IndexFunc(notes, func(note model.Note) bool {
		return note.Id == int(noteID)
	})
	return &remindables_api.GetNoteResponse{
		Id:             int32(notes[idx].Id),
		Name:           notes[idx].Name,
		Description:    notes[idx].Description,
		AlarmTimeStamp: timestamppb.New(notes[idx].AlarmTimeStamp),
	}, nil
}

// PostNewTask implements remindables_api.RemindablesServiceClient.
func (s *server) PostNewTask(
	ctx context.Context,
	userRequest *remindables_api.PostNewTaskRequest,
) (*remindables_api.PostNewTaskResponse, error) {
	name := userRequest.GetName()
	description := userRequest.GetDescription()
	dueDate := userRequest.GetDueDate().AsTime()
	task := *repository.PostNewTask(name, description, dueDate)
	if task.Id == 0 {
		return &remindables_api.PostNewTaskResponse{}, fmt.Errorf("ошибка создания новой задачи")
	}
	return &remindables_api.PostNewTaskResponse{
		Id:            int32(task.Id),
		Name:          task.Name,
		Description:   task.Description,
		InitTimeStamp: timestamppb.New(task.InitTimeStamp),
		DueDate:       timestamppb.New(task.DueDate),
		Status:        task.Status,
	}, nil
}

// PostNewNote implements remindables_api.RemindablesServiceClient.
func (s *server) PostNewNote(
	ctx context.Context,
	userRequest *remindables_api.PostNewNoteRequest,
) (*remindables_api.PostNewNoteResponse, error) {
	name := userRequest.GetName()
	description := userRequest.GetDescription()
	alarmTimeStamp := userRequest.GetAlarmTimeStamp().AsTime()
	note := *repository.PostNewNote(name, description, alarmTimeStamp)
	if note.Id == 0 {
		return &remindables_api.PostNewNoteResponse{}, fmt.Errorf("ошибка создания новой заметки")
	}
	return &remindables_api.PostNewNoteResponse{
		Id:             int32(note.Id),
		Name:           note.Name,
		Description:    note.Description,
		AlarmTimeStamp: timestamppb.New(note.AlarmTimeStamp),
	}, nil
}

// PutTaskById implements remindables_api.RemindablesServiceClient.
func (s *server) PutTaskById(
	ctx context.Context,
	userRequest *remindables_api.PutTaskRequest,
) (*remindables_api.PutTaskResponse, error) {
	id := userRequest.GetId()
	name := userRequest.GetName()
	description := userRequest.GetDescription()
	dueDate := userRequest.GetDueDate().AsTime()
	task := *repository.PutTaskById(id, name, description, dueDate)
	if task.Id == 0 {
		return &remindables_api.PutTaskResponse{}, fmt.Errorf("ошибка изменения задачи")
	}
	return &remindables_api.PutTaskResponse{
		Id:            int32(task.Id),
		Name:          task.Name,
		Description:   task.Description,
		InitTimeStamp: timestamppb.New(task.InitTimeStamp),
		DueDate:       timestamppb.New(task.DueDate),
		Status:        task.Status,
	}, nil
}

// PutNoteById implements remindables_api.RemindablesServiceClient.
func (s *server) PutNoteById(
	ctx context.Context,
	userRequest *remindables_api.PutNoteRequest,
) (*remindables_api.PutNoteResponse, error) {
	id := userRequest.GetId()
	name := userRequest.GetName()
	description := userRequest.GetDescription()
	alarmTimeStamp := userRequest.GetAlarmTimeStamp().AsTime()
	note := *repository.PutNoteById(id, name, description, alarmTimeStamp)
	if note.Id == 0 {
		return &remindables_api.PutNoteResponse{}, fmt.Errorf("ошибка изменения заметки")
	}
	return &remindables_api.PutNoteResponse{
		Id:             int32(note.Id),
		Name:           note.Name,
		Description:    note.Description,
		AlarmTimeStamp: timestamppb.New(note.AlarmTimeStamp),
	}, nil
}

// DeleteTaskById implements remindables_api.RemindablesServiceClient.
func (s *server) DeleteTaskById(
	ctx context.Context,
	userRequest *remindables_api.DeleteTaskRequest,
) (*remindables_api.DeleteTaskResponse, error) {
	id := userRequest.GetId()
	task := *repository.DeleteTaskById(id)
	if task.Id == 0 {
		return &remindables_api.DeleteTaskResponse{}, fmt.Errorf("ошибка удаления задачи")
	}
	return &remindables_api.DeleteTaskResponse{
		Id:            int32(task.Id),
		Name:          task.Name,
		Description:   task.Description,
		InitTimeStamp: timestamppb.New(task.InitTimeStamp),
		DueDate:       timestamppb.New(task.DueDate),
		Status:        task.Status,
	}, nil
}

// DeleteNoteById implements remindables_api.RemindablesServiceClient.
func (s *server) DeleteNoteById(
	ctx context.Context,
	userRequest *remindables_api.DeleteNoteRequest,
) (*remindables_api.DeleteNoteResponse, error) {
	id := userRequest.GetId()
	note := *repository.DeleteNoteById(id)
	if note.Id == 0 {
		return &remindables_api.DeleteNoteResponse{}, fmt.Errorf("ошибка удаления заметки")
	}
	return &remindables_api.DeleteNoteResponse{
		Id:             int32(note.Id),
		Name:           note.Name,
		Description:    note.Description,
		AlarmTimeStamp: timestamppb.New(note.AlarmTimeStamp),
	}, nil
}

func createFiles(fileNames ...string) {
	wg := sync.WaitGroup{}
	for _, fileName := range fileNames {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			// O_EXCL - used with O_CREATE, file must not exist.
			file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				// если файл существует
				if errors.Is(err, os.ErrExist) {
					return
				}
				// при всяких других ошибках
				panic(err)
			}
			fmt.Printf("Файл %s создан.\n", fileName)
			defer file.Close()
		}(fileName)
	}
	wg.Wait()
	fmt.Println()
}

func main() {
	// создание json-файлов для хранения задач и заметок
	createFiles("tasks.json", "notes.json")

	wg := sync.WaitGroup{}
	errsTasks := make(chan error, 1)
	errsNotes := make(chan error, 1)

	// заполнение срезов задач и заметок соответствующими данными из json-файлов
	wg.Add(2)
	go repository.FillTasksFromJSON(&wg, errsTasks)
	go repository.FillNotesFromJSON(&wg, errsNotes)

	func() {
		wg.Wait()
		close(errsTasks)
		close(errsNotes)
	}()

	if err := <-errsTasks; err != nil {
		panic(fmt.Sprintf("ошибка наполнения repository.Tasks из файла tasks.json: %v", err))
	}
	if err := <-errsNotes; err != nil {
		panic(fmt.Sprintf("ошибка наполнения repository.Notes из файла notes.json: %v", err))
	}

	lis, err := net.Listen("tcp", "localhost:5001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			loggingInterceptor,
		),
		grpc.StreamInterceptor(
			loggingStreamInterceptor,
		),
	)
	remindables_api.RegisterRemindablesServiceServer(s, &server{})

	reflection.Register(s)

	log.Println("Server is running at :5001")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

func loggingStreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()

	err := handler(srv, ss)

	log.Printf(
		"[gRPC] method=%s error=%v duration=%s",
		info.FullMethod,
		err,
		time.Since(start),
	)

	return err
}

func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()

	resp, err = handler(ctx, req)

	st, _ := status.FromError(err)

	var reqJSON, respJSON string

	if m, ok := req.(proto.Message); ok {
		b, _ := protojson.Marshal(m)
		reqJSON = string(b)
	} else {
		reqJSON = "<non-proto request>"
	}

	if m, ok := resp.(proto.Message); ok && resp != nil {
		b, _ := protojson.Marshal(m)
		respJSON = string(b)
	} else {
		respJSON = "<non-proto response or nil>"
	}

	log.Printf(
		"[gRPC] method=%s status=%s error=%v duration=%s request=%s response=%s",
		info.FullMethod,
		st.Code(),
		err,
		time.Since(start),
		reqJSON,
		respJSON,
	)

	return resp, err
}
