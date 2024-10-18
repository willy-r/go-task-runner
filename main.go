package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskService struct {
	DB          *sql.DB
	TaskChannel chan Task
}

func (taskService *TaskService) AddTask(task Task) (int64, error) {
	query := "INSERT INTO tasks (title, description, status, created_at) VALUES (?, ?, ?, ?)"
	result, err := taskService.DB.Exec(query, task.Title, task.Description, task.Status, task.CreatedAt)
	insertedId, _ := result.LastInsertId()
	return insertedId, err
}

func (taskService *TaskService) UpdateTaskStatus(task Task) error {
	query := "UPDATE tasks SET status = ? WHERE id = ?"
	_, err := taskService.DB.Exec(query, task.Status, task.ID)
	return err
}

func (taskService *TaskService) ListTasks() ([]Task, error) {
	query := "SELECT * FROM tasks"
	rows, err := taskService.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (taskService *TaskService) ProcessTasks() {
	for task := range taskService.TaskChannel {
		log.Printf("Processing task: %s", task.Title)
		time.Sleep(5 * time.Second)
		task.Status = "COMPLETED"
		taskService.UpdateTaskStatus(task)
		log.Printf("Task %s processed", task.Title)
	}
}

func (taskService *TaskService) HandleCreateTask(writer http.ResponseWriter, request *http.Request) {
	var task Task
	err := json.NewDecoder(request.Body).Decode(&task)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	task.Status = "PENDING"
	task.CreatedAt = time.Now()
	insertedId, err := taskService.AddTask(task)
	if err != nil {
		http.Error(writer, "Error addding task, try again later", http.StatusInternalServerError)
		return
	}
	task.ID = int(insertedId)
	taskService.TaskChannel <- task // Task go to processing, putting it in a channel
	writer.WriteHeader(http.StatusCreated)
}

func (taskService *TaskService) HandleListTasks(writer http.ResponseWriter, request *http.Request) {
	tasks, err := taskService.ListTasks()
	if err != nil {
		http.Error(writer, "Error listing tasks", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tasks)
}

func main() {
	db, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	taskChannel := make(chan Task)

	taskService := TaskService{
		DB:          db,
		TaskChannel: taskChannel,
	}

	go taskService.ProcessTasks()

	http.HandleFunc("/tasks", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			taskService.HandleCreateTask(writer, request)
		} else if request.Method == http.MethodGet {
			taskService.HandleListTasks(writer, request)
		} else {
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server running on port :8081")
	http.ListenAndServe(":8081", nil)
}
