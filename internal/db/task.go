package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tracker/internal/constants"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

var (
    ErrSelectTasks = errors.New("ошибка получения задач")
    ErrReadResult  = errors.New("не удалось считать результат")
    ErrDeleteTask  = errors.New("не удалось удалить задачу")
)

func AddTask(task *Task) (int64, error) {
    var id int64
    // определите запрос 
    query := `INSERT INTO scheduler(date,title,comment,repeat) VALUES ($1,$2,$3,$4)`
    res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
    if err == nil {
        id, err = res.LastInsertId()
    }
    return id, err
}

func Tasks(limit int, search string) ([]*Task, error) {
    var (
        res *sql.Rows
        err error
    )
    
    query := "SELECT id,date,title,comment,repeat FROM scheduler LIMIT $1"
    if search != "" {
        value, isDate := checkSearch(search)
        if !isDate {
            value = "%" + value + "%"
            query = "SELECT date,title,comment,repeat FROM scheduler WHERE title LIKE $2 OR comment LIKE $2 ORDER BY date LIMIT $1"
        } else {
            query = "SELECT date,title,comment,repeat FROM scheduler WHERE date = $2 LIMIT $1"
        }
        res, err = db.Query(query, limit, value)
    } else {
        res, err = db.Query(query, limit) 
    }

    if err != nil {
        return nil, fmt.Errorf("%w - %w", ErrSelectTasks, err) // ErrSelectTasks
    }

    var result = make([]*Task, 0, limit)
    for res.Next() {
        var task Task
        err = res.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
        if err != nil {
            return nil, ErrReadResult
        }
        result = append(result, &task)
    }

    if res.Err() != nil {
        return nil, ErrReadResult
    }

    return result, nil
}

// true - значит поиск по дате должен быть
func checkSearch(search string) (string, bool) {
    date, err := time.Parse(constants.TimeParsePattern, search)
    if err == nil {
        return date.Format(constants.TimeFormat), true
    }
    return search, false
}

func GetTask(id string) (*Task, error) {
    query := "SELECT * FROM scheduler WHERE id=$1"
    var task Task
    err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
    if err != nil {
        return nil, ErrSelectTasks
    }
    return &task, nil
}

func UpdateTask(task *Task) error {
    // параметры пропущены, не забудьте указать WHERE
    query := `UPDATE scheduler SET date=$2, title=$3, comment=$4, repeat=$5 WHERE id=$1`
    res, err := db.Exec(query, task.ID, task.Date, task.Title, task.Comment, task.Repeat)
    if err != nil {
        return err
    }
    // метод RowsAffected() возвращает количество записей к которым 
    // был применена SQL команда 
    count, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if count == 0 {
        return fmt.Errorf(`некорректный id для обновления задачи`)
    }
    return nil
}

func DeleteTask(id string) error {
    query := "DELETE FROM scheduler WHERE id=$1;"
    res, err := db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("ошибка при выполнении действия для задачи с id %s: %w",id, ErrDeleteTask)
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("что-то поломалось в базе")
    }
    if rows == 0 {
        return fmt.Errorf("задачи с заданным id не существует")
    }

    return nil
}