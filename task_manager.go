package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

// Color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

type TaskManager struct {
	Title       string
	Description string
	Deadline    string
}

func load() []TaskManager {
	path, _ := os.UserHomeDir() // gets user home folder
	path += "/.task.mine"
	f, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if len(f) == 0 {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(string(f))
	if err != nil {
		return nil
	}
	var tasks []TaskManager
	json.Unmarshal(data, &tasks)
	return tasks
}

func save(task []TaskManager) {
	path, _ := os.UserHomeDir() // get user home folder
	path += "/.task.mine"
	data, err := json.Marshal(task)
	if err != nil {
		fmt.Println(Red + "Error saving task list: " + err.Error() + Reset)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(Red + "Error saving task list: " + err.Error() + Reset)
		return
	}
	defer f.Close()
	f.WriteString(encoded)
	fmt.Println(Green + "Task list saved successfully." + Reset)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	task := load()
	tmpl.Execute(w, struct {
		Tasks []TaskManager
	}{
		Tasks: task,
	})
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		item := r.FormValue("item")
		description := r.FormValue("description")
		deadline := r.FormValue("deadline")
		task := load()
		task = append(task, TaskManager{Title: item, Description: description, Deadline: deadline})
		save(task)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		toDelete := r.Form["delete"]
		task := load()
		var newTodo []TaskManager
		for i, item := range task {
			found := false
			for _, indexStr := range toDelete {
				index, err := strconv.Atoi(indexStr)
				if err == nil && index == i {
					found = true
					break
				}
			}
			if !found {
				newTodo = append(newTodo, item)
			}
		}
		save(newTodo)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		save(nil)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		index, err := strconv.Atoi(r.FormValue("index"))
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		title := r.FormValue("item")
		description := r.FormValue("description")
		deadline := r.FormValue("deadline")
		task := load()
		if index >= 0 && index < len(task) {
			task[index] = TaskManager{Title: title, Description: description, Deadline: deadline}
			save(task)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/clear", clearHandler)
	http.HandleFunc("/update", updateHandler)

	fmt.Println("Starting server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(Red + "Failed to start server: " + err.Error() + Reset)
	}
}
