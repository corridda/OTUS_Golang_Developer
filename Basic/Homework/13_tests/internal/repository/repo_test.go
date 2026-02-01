package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Искусственное заполенение списков задач и заметок из `n` элементов
func createList() {
	length := len(Tasks)
	CreateNewRemindable(
		fmt.Sprintf("taskName %d", length+1),
		fmt.Sprintf("taskDescr %d", length+1),
		"01.02.2026",
		true)
	length = len(Notes)
	CreateNewRemindable(
		fmt.Sprintf("noteName %d", length+1),
		fmt.Sprintf("noteDescr %d", length+1),
		"01.02.2026 20:00",
		false,
	)
}

// Выделение общих ресурсов для всех тестов
func TestMain(m *testing.M) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Создать json-файлы для хранения задач и заметок для тестовых целей
	fileNames := []string{"tasks.json", "notes.json"}
	wg := sync.WaitGroup{}
	for _, fileName := range fileNames {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			file, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Файл %s создан.\n", fileName)
			defer func() {
				if err := file.Close(); err != nil {
					panic(err)
				}
			}()
		}(fileName)
	}
	wg.Wait()
	fmt.Println()

	createList()

	errsTasks := make(chan error, 1)
	errsNotes := make(chan error, 1)

	// заполнение срезов задач и заметок соответствующими данными из json-файлов
	wg.Add(2)
	go FillTasksFromJSON(&wg, errsTasks)
	go FillNotesFromJSON(&wg, errsNotes)

	go func() {
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

	// Запустить все тесты
	exitCode := m.Run()

	// Удалить ранее созданные json-файлы
	for _, fileName := range []string{"tasks.json", "notes.json"} {
		if err := os.Remove(fileName); err != nil {
			panic(err)
		}
		fmt.Printf("Файл %s удален.\n", fileName)
	}

	// Выход из программы с результатом выполнения тестов
	os.Exit(exitCode)
}

func TestGetNotes(t *testing.T) {
	t.Run("notes getting success", func(t *testing.T) {
		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/notes/items", nil)
		ctx.Request = req

		// 4. Call the handler function directly
		GetNotes(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetTasksById(t *testing.T) {
	t.Run("success with getting a task by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/tasks/item/id/?id=1", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 4. Call the handler function directly
		GetTasksById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("fail with getting a task by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/tasks/item/id/?id=10", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		GetTasksById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Задачи с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("getting a task by id bad request", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/tasks/item/id", nil)
		ctx.Request = req

		// 4. Call the handler function directly
		GetTasksById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusBadRequest, w.Code)
		expectedBody := `{"error":"Key: 'RemindableId.Id' Error:Field validation for 'Id' failed on the 'required' tag"}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestGetNotesById(t *testing.T) {
	t.Run("success with getting a note by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/notes/item/id/?id=1", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 4. Call the handler function directly
		GetNotesById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody := `{"id":1,"name":"noteName 1","description":"noteDescr 1","alarmTimeStamp":"2026-02-01T20:00:00Z"}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("fail with getting a note by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/notes/item/id/?id=10", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		GetNotesById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Заметки с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("getting a note by id bad request", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/notes/item/id", nil)
		ctx.Request = req

		// 4. Call the handler function directly
		GetTasksById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusBadRequest, w.Code)
		expectedBody := `{"error":"Key: 'RemindableId.Id' Error:Field validation for 'Id' failed on the 'required' tag"}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestGetTasks(t *testing.T) {
	t.Run("tasks getting success", func(t *testing.T) {
		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodGet, "/api/tasks/items", nil)
		ctx.Request = req

		// 4. Call the handler function directly
		GetTasks(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestPostNewTask(t *testing.T) {
	t.Run("new task creation success", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description" binding:"required"`
			DueDate     string `json:"dueDate" binding:"required"`
		}{
			Name:        "taskName 2",
			Description: "taskDescr 2",
			DueDate:     "03.02.2026",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPost, "/api/tasks/item", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		// 5. Call the handler function directly
		PostNewTask(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("new task creation fail", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description" binding:"required"`
			DueDate     string `json:"dueDate" binding:"required"`
		}{
			Name:        "taskName 3",
			Description: "taskDescr 3",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPost, "/api/tasks/item", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		// 5. Call the handler function directly
		PostNewTask(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusBadRequest, w.Code)
		expectedBody := `{"error":"Key: 'NewTask.DueDate' Error:Field validation for 'DueDate' failed on the 'required' tag"}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestPostNewNote(t *testing.T) {
	t.Run("new note creation success", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name           string `json:"name" binding:"required"`
			Description    string `json:"description" binding:"required"`
			AlarmTimeStamp string `json:"alarmTime" binding:"required"`
		}{
			Name:           "noteName 2",
			Description:    "noteDescr 2",
			AlarmTimeStamp: "03.02.2026 20:00",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPost, "/api/notes/item", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		// 5. Call the handler function directly
		PostNewNote(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody := `{"Создана новая заметка":{"id":2,"name":"noteName 2","description":"noteDescr 2","alarmTimeStamp":"2026-02-03T20:00:00Z"}}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("new note creation fail", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name           string `json:"name" binding:"required"`
			Description    string `json:"description" binding:"required"`
			AlarmTimeStamp string `json:"alarmTime" binding:"required"`
		}{
			Name:        "noteName 3",
			Description: "noteDescr 3",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPost, "/api/notes/item", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		// 5. Call the handler function directly
		PostNewNote(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusBadRequest, w.Code)
		expectedBody := `{"error":"Key: 'NewNote.AlarmTimeStamp' Error:Field validation for 'AlarmTimeStamp' failed on the 'required' tag"}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestPutTaskById(t *testing.T) {
	t.Run("change task by id success", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			DueDate     string `json:"dueDate"`
		}{
			Name:        "new taskName 1",
			Description: "new taskDescr 1",
			DueDate:     "01.01.2030",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPut, "/api/tasks/item/id/?id=1", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 5. Call the handler function directly
		PutTaskById(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		containBody := `{"Изменена задача":{"id":1,"name":"new taskName 1","description":"new taskDescr 1",`
		assert.Contains(t, w.Body.String(), containBody)
	})

	t.Run("change task by id failure", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			DueDate     string `json:"dueDate"`
		}{
			Name:        "new taskName 1",
			Description: "new taskDescr 1",
			DueDate:     "01.01.2030",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPut, "/api/notes/item/id/?id=10", bytes.NewBuffer(jsonData))
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		PutTaskById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Задачи с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestPutNoteById(t *testing.T) {
	t.Run("change note by id success", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name           string `json:"name"`
			Description    string `json:"description"`
			AlarmTimeStamp string `json:"alarmTimeStamp"`
		}{
			Name:           "new noteName 1",
			Description:    "new noteDescr 1",
			AlarmTimeStamp: "01.01.2030 15:25",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPut, "/api/notes/item/id/?id=1", bytes.NewBuffer(jsonData))

		// 4. Crucially, set the "Content-Type" header to "application/json"
		// so your handler knows how to interpret the body.
		req.Header.Set("Content-Type", "application/json")

		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 5. Call the handler function directly
		PutNoteById(ctx)

		// 6. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody := `{"Изменена заметка":{"id":1,"name":"new noteName 1","description":"new noteDescr 1","alarmTimeStamp":"2030-01-01T15:25:00Z"}}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("change note by id failure", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		data := struct {
			Name           string `json:"name"`
			Description    string `json:"description"`
			AlarmTimeStamp string `json:"alarmTimeStamp"`
		}{
			Name:           "new noteName 10",
			Description:    "new noteDescr 10",
			AlarmTimeStamp: "01.01.2030 15:25",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodPut, "/api/notes/item/id/?id=10", bytes.NewBuffer(jsonData))
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		PutNoteById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Заметки с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestDeleteTaskById(t *testing.T) {
	t.Run("success with deleting a task by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodDelete, "/api/tasks/item/id/?id=1", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 4. Call the handler function directly
		DeleteTaskById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		containBody := `{"Удалена задача":{"id":1,"name":"new taskName 1","description":"new taskDescr 1",`
		assert.Contains(t, w.Body.String(), containBody)
	})

	t.Run("fail with deleting a task by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodDelete, "/api/tasks/item/id/?id=10", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		DeleteTaskById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Задачи с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}

func TestDeleteNoteById(t *testing.T) {
	t.Run("success with deleting a note by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodDelete, "/api/notes/item/id/?id=1", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "1"},
		}

		// 4. Call the handler function directly
		DeleteNoteById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody := `{"Удалена заметка":{"id":1,"name":"new noteName 1","description":"new noteDescr 1","alarmTimeStamp":"2030-01-01T15:25:00Z"}}`
		assert.Equal(t, expectedBody, w.Body.String())
	})

	t.Run("fail with deleting a note by id", func(t *testing.T) {
		t.Parallel()

		// 1. Create a mock response recorder
		w := httptest.NewRecorder()

		// 2. Create a mock context and engine (engine is needed for params to work correctly)
		ctx, _ := gin.CreateTestContext(w)

		// 3. Mock the request, setting the URL path
		req, _ := http.NewRequest(http.MethodDelete, "/api/notes/item/id/?id=10", nil)
		ctx.Request = req

		ctx.Params = []gin.Param{
			{Key: "id", Value: "10"},
		}

		// 4. Call the handler function directly
		DeleteNoteById(ctx)

		// 5. Assert the results
		assert.Equal(t, http.StatusNotFound, w.Code)
		expectedBody := `{"NotFound":"Заметки с id=10 не существует."}`
		assert.Equal(t, expectedBody, w.Body.String())
	})
}
