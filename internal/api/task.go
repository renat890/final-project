package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"tracker/internal/constants"
	"tracker/internal/db"
)

var (
	ErrEmptyTitleParam  = errors.New("поле Title пустое")
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	// десериализация JSON
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusBadRequest)
		return
	}

	// проверка поля task.Title на пустоту
	if task.Title == "" {
		writeResponse(w, "error", ErrEmptyTitleParam.Error(), http.StatusBadRequest)
		return
	}

	// проверка на корректность полученное значение task.Date
	if err := checkDate(&task); err != nil {
		writeResponse(w, "error", err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}

	writeResponse(w, "id", fmt.Sprint(id), http.StatusOK)

}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	task, err := db.GetTask(id)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, &task)
}

func modifyTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusBadRequest)
		return
	}

	// проверка поля task.Title на пустоту
	if task.Title == "" {
		writeResponse(w, "error", ErrEmptyTitleParam.Error(), http.StatusBadRequest)
		return
	}

	// проверка на корректность полученное значение task.Date
	if err := checkDate(&task); err != nil {
		writeResponse(w, "error", err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateTask(&task)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, struct{}{})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		search := r.FormValue("search")
		tasks, err := db.Tasks(50, search)
		if err != nil {
			writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
			return
		}
		writeJson(w, TasksResp{
			Tasks: tasks,
		})
	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	}
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponse(w, "error", ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	// получение и проверка параметра id
	id := r.FormValue("id")
	if id == "" {
		writeResponse(w, "error", ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}
	// получение деталей задачи
	task, err := db.GetTask(id)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	// проверка repeat
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
			return
		}
		writeJson(w, struct{}{})
		return
	}
	// для периодической задачи
	nDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	task.Date = nDate
	if err := db.UpdateTask(task); err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, struct{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeResponse(w, "error", ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, struct{}{})

}

func writeJson(w http.ResponseWriter, body any) {
	jBody, err := json.Marshal(body)
	if err != nil {
		writeResponse(w, "error", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jBody)
}

func writeResponse(w http.ResponseWriter, typeR string, value string, status int) {
	var body []byte

	switch typeR {
	case "id":
		type rBody struct {
			ID string `json:"id"`
		}
		body, _ = json.Marshal(rBody{ID: value})
	case "error":
		type rBody struct {
			Error string `json:"error"`
		}
		body, _ = json.Marshal(rBody{Error: value})
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_, err := w.Write(body)
	if err != nil {
		log.Printf("не удалось записать ответ: %s", err.Error())
	}
}

func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(constants.TimeFormat)
	}

	t, err := time.Parse(constants.TimeFormat, task.Date)
	if err != nil {
		return fmt.Errorf("не могу распарсить время: %w", err)
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	  // если сегодня (now) больше task.Date (t)
    if afterNow(now, t) {
        if len(task.Repeat) == 0 {
            // если правила повторения нет, то берём сегодняшнее число
            task.Date = now.Format(constants.TimeFormat)
        } else {
            // в противном случае, берём вычисленную ранее следующую дату
            task.Date = next
        }
    }

	return nil
}