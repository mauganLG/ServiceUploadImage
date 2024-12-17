package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/tutti-ch/backend-coding-task-template/image"
	"github.com/tutti-ch/backend-coding-task-template/response"
	"github.com/tutti-ch/backend-coding-task-template/worker"
)

const (
	// File Size Limit 8192 Kilobytes
	maxUploadSize = 8192 * 1024
)

// pool of worker
var workers *worker.Workers

func main() {

	var fsDirectory string

	if fsDirectory = os.Getenv("BASE_PATH"); fsDirectory == "" {
		fmt.Fprintf(os.Stderr, "the `BASE_PATH` environment variable not supplied")
		return
	}

	if _, err := os.Stat(fsDirectory); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "the path in `BASE_PATH` %s not exist", fsDirectory)
		return
	}

	nbWorkers, err := strconv.Atoi(os.Getenv("WORKERS"))
	if err != nil || nbWorkers < 1 {
		nbWorkers = runtime.NumCPU()
	}

	port := "8888"
	var sb strings.Builder

	log.Printf("Server $BASE_PATH : %s\n", fsDirectory)
	log.Printf("Server $WORKERS number of workers : %d\n", nbWorkers)

	sb.WriteString(":")
	sb.WriteString(port)

	httpServer := &http.Server{
		Addr: sb.String(),
	}

	http.Handle("/upload", MakeHandler(fsDirectory, nbWorkers))

	go func() {
		log.Printf("Server listening on port : %s", port)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
		log.Println("Server not accepting new connections")

	}()

	// Graceful shutdown that catch the signal associated
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithCancel(context.Background())
	shutdownRelease()
	workers.Shutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server shutdown")

}

// Handle files of jpeg format less than 8192 Kib and process them throught defined workers
func MakeHandler(fsDirectory string, poolSize int) http.Handler {
	workers = worker.NewWorkers(poolSize)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		request.Body = http.MaxBytesReader(writer, request.Body, maxUploadSize)

		if err := request.ParseMultipartForm(maxUploadSize); err != nil {
			response.ErrorHandler(writer, "file is bigger than 8192 Kilobytes", http.StatusRequestEntityTooLarge)
			log.Printf("file is bigger than 8192 Kilobytes : %s\n", err.Error())
			return
		}

		file, _, err := request.FormFile("image")

		if err != nil {
			response.ErrorHandler(writer, "file is not an image type", http.StatusBadRequest)
			log.Printf("file is not an image type: %s\n", err.Error())
			return
		}
		defer file.Close()

		buffer := make([]byte, 512)
		if _, err := file.Read(buffer); err != nil {
			response.ErrorHandler(writer, "failed to read file", http.StatusInternalServerError)
			log.Printf("failed to read file: %s\n", err.Error())
			return
		}
		file.Seek(0, io.SeekStart)
		if http.DetectContentType(buffer) != "image/jpeg" {
			response.ErrorHandler(writer, "file is not a JPEG image", http.StatusBadRequest)
			log.Println("file is not a JPEG image")
			return
		}

		uuid := uuid.NewString()
		filePath := fmt.Sprintf("%s/%s.jpg", fsDirectory, uuid)
		if workers.Task(func() {

			if err := image.ProcessImage(filePath, file); err != nil {
				response.ErrorHandler(writer, err.Error(), http.StatusInternalServerError)
				log.Printf("Error Process Images : %s\n", err.Error())
			}
		}) {
			response.OK(writer, uuid)
			log.Printf("File %s saved\n", filePath)
			return
		} else {
			response.ErrorHandler(writer, "too many requests", http.StatusTooManyRequests)
			log.Println("too many requests")
			return
		}
	})
}
