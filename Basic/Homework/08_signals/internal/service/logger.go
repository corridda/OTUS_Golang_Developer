package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/corridda/OTUS_Golang_Developer/Basic/Homework/08_signals/internal/repository"
)

var taskHashes [][32]byte
var noteHashes [][32]byte

func LogRemidables(ctx context.Context, ticker *time.Ticker, mutex *sync.RWMutex) {
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			fmt.Println("\nЛоггер завершил процесс логирования.")
			return
		case <-ticker.C:
			mutex.RLock()

			// Проверить на наличие новых задач и логировать их в консоль, сохраняя хэш
			for _, task := range repository.Tasks {
				taskHash := sha256.Sum256([]byte(task.String()))
				// fmt.Printf("taskHash: %x\n", taskHash)
				if !slices.Contains(taskHashes, taskHash) {
					taskHashes = append(taskHashes, taskHash)
					fmt.Println("*** Добавлена новая задача ***")
					fmt.Println(task.String())
				}
			}

			// Проверить на наличие новых заметок и логировать их в консоль, сохраняя хэш
			for _, note := range repository.Notes {
				noteHash := sha256.Sum256([]byte(note.String()))
				if !slices.Contains(noteHashes, noteHash) {
					noteHashes = append(noteHashes, noteHash)
					fmt.Println("*** Добавлена новая заметка ***")
					fmt.Println(note.String())
				}
			}

			mutex.RUnlock()
		}
	}
}
