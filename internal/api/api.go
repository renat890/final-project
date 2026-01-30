package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrMethodNotAllowed = errors.New("для данного обработчика выбран неправильный метод")
)

func Init() {
    http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		now := r.FormValue("now")
		date := r.FormValue("date")
		repeat := r.FormValue("repeat")

		// now и date должен быть заполнен
		if now == "" || date == "" {
			http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
			return
		}

		nowT, err := time.Parse(FORMAT, now)
		if err != nil {
			http.Error(w, fmt.Errorf("не могу распарсить now, ошибка: %w", err).Error(), http.StatusBadRequest)
			return
		}

		nextDate, err := NextDate(nowT, date, repeat)
		if err != nil {
			http.Error(w, fmt.Errorf("не могу вычислить следующую дату: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(nextDate))
	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	}
}



func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w,r)
	case http.MethodPut:
		modifyTaskHandler(w,r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	}
}

