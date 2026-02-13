package main

import (
	"fmt"
	"strings"
)

// Task represents a single todo item
type Task struct {
	ID          int
	Description string
	Completed   bool
}

// TodoList manages a collection of tasks
type TodoList struct {
	tasks  []Task
	nextID int
}

// NewTodoList creates a new todo list
func NewTodoList() *TodoList {
	return &TodoList{
		tasks:  make([]Task, 0),
		nextID: 1,
	}
}

// AddTask adds a new task to the list
func (tl *TodoList) AddTask(description string) {
	task := Task{
		ID:          tl.nextID,
		Description: description,
		Completed:   false,
	}
	tl.tasks = append(tl.tasks, task)
	tl.nextID++
	fmt.Printf("✅ Added task: %s\n", description)
}

// ListTasks displays all tasks
func (tl *TodoList) ListTasks() {
	if len(tl.tasks) == 0 {
		fmt.Println("📝 No tasks yet! Add some tasks to get started.")
		return
	}

	fmt.Println("\n📋 Your Tasks:")
	fmt.Println(strings.Repeat("-", 40))

	for _, task := range tl.tasks {
		status := "❌"
		if task.Completed {
			status = "✅"
		}
		fmt.Printf("%s [%d] %s\n", status, task.ID, task.Description)
	}
	fmt.Println(strings.Repeat("-", 40))
}

// CompleteTask marks a task as completed
func (tl *TodoList) CompleteTask(id int) {
	for i := range tl.tasks {
		if tl.tasks[i].ID == id {
			if tl.tasks[i].Completed {
				fmt.Printf("🔄 Task %d is already completed!\n", id)
				return
			}
			tl.tasks[i].Completed = true
			fmt.Printf("🎉 Completed task: %s\n", tl.tasks[i].Description)
			return
		}
	}
	fmt.Printf("❓ Task with ID %d not found!\n", id)
}

// DeleteTask removes a task from the list
func (tl *TodoList) DeleteTask(id int) {
	for i, task := range tl.tasks {
		if task.ID == id {
			// Remove task from slice
			tl.tasks = append(tl.tasks[:i], tl.tasks[i+1:]...)
			fmt.Printf("🗑️ Deleted task: %s\n", task.Description)
			return
		}
	}
	fmt.Printf("❓ Task with ID %d not found!\n", id)
}

// GetStats returns statistics about the tasks
func (tl *TodoList) GetStats() (total, completed, pending int) {
	total = len(tl.tasks)
	for _, task := range tl.tasks {
		if task.Completed {
			completed++
		} else {
			pending++
		}
	}
	return
}

func main() {
	fmt.Println("📝 Welcome to Go Todo List Manager!")
	fmt.Println("Commands: add, list, complete, delete, stats, help, quit")

	todoList := NewTodoList()

	// Add some sample tasks
	todoList.AddTask("Learn Go basics")
	todoList.AddTask("Build a todo app")
	todoList.AddTask("Practice Go structs and methods")

	for {
		fmt.Print("\n> ")
		var command string
		fmt.Scanln(&command)

		switch strings.ToLower(command) {
		case "add":
			fmt.Print("Enter task description: ")
			var description string
			fmt.Scanln(&description)
			if description != "" {
				todoList.AddTask(description)
			} else {
				fmt.Println("❌ Task description cannot be empty!")
			}

		case "list":
			todoList.ListTasks()

		case "complete":
			fmt.Print("Enter task ID to complete: ")
			var id int
			_, err := fmt.Scanf("%d", &id)
			if err != nil {
				fmt.Println("❌ Please enter a valid task ID!")
				continue
			}
			todoList.CompleteTask(id)

		case "delete":
			fmt.Print("Enter task ID to delete: ")
			var id int
			_, err := fmt.Scanf("%d", &id)
			if err != nil {
				fmt.Println("❌ Please enter a valid task ID!")
				continue
			}
			todoList.DeleteTask(id)

		case "stats":
			total, completed, pending := todoList.GetStats()
			fmt.Printf("\n📊 Statistics:\n")
			fmt.Printf("   Total tasks: %d\n", total)
			fmt.Printf("   Completed: %d\n", completed)
			fmt.Printf("   Pending: %d\n", pending)
			if total > 0 {
				fmt.Printf("   Progress: %.1f%%\n", float64(completed)/float64(total)*100)
			}

		case "help":
			fmt.Println("\n📖 Available Commands:")
			fmt.Println("   add      - Add a new task")
			fmt.Println("   list     - Show all tasks")
			fmt.Println("   complete - Mark a task as completed")
			fmt.Println("   delete   - Remove a task")
			fmt.Println("   stats    - Show task statistics")
			fmt.Println("   help     - Show this help message")
			fmt.Println("   quit     - Exit the program")

		case "quit", "exit":
			fmt.Println("👋 Thanks for using Go Todo List Manager!")
			return

		default:
			fmt.Printf("❓ Unknown command: %s\n", command)
			fmt.Println("Type 'help' to see available commands.")
		}
	}
}
