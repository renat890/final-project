package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tracker/internal/constants"
)

var (
	ErrInvalidMonthParam = errors.New("невалидное значение для дня месяца")
	ErrIvalidWeekParam   = errors.New("невалидное значение для дня недели")
	ErrEmptyParam        = errors.New("пустое поле now или date, должно быть заполнено")
	ErrInvalidInterval   = errors.New("превышен максимально допустимый интервал")
	ErrEmptyDay          = errors.New("не указан интервал в днях")
	ErrEmptyRepeat       = errors.New("в параметре repeat — пустая строка")
	ErrInvalidFormat     = errors.New("указан неверный формат repeat")
	
)

func afterNow(date, now time.Time) bool {
	dateN := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, now.Location())
	nowN := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return dateN.After(nowN)
}

// проверка дня недели, также корректность значения параметра week
func isDayWeek(date time.Time, daysOfWeek []string) (bool, error) {
	var minDay = 1
	var maxDay = 7

	// хитро устанавливаем на русский лад
	// воскресенье со значением 0 преврщается в воскресенье со значением 7
	russDay := int(date.Weekday())
	if date.Weekday() == 0 {
		russDay = 7
	}
	

	for _, val := range daysOfWeek {
		intVal, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return false, fmt.Errorf("ошибка парсинга: %w", ErrIvalidWeekParam)
		}
		if int(intVal) > maxDay || int(intVal) < minDay {
			return false, fmt.Errorf("невалдиное значение дня недели: %w", ErrIvalidWeekParam)
		}
		if russDay == int(intVal) {
			return true, nil
		}
	}

	return false, nil
}

func isDayMonth(date time.Time, days [32]bool, months [13]bool) (bool, error) {
	day   := date.Day()
	month := date.Month()

	_ = day
	_ = month

	return false, ErrInvalidMonthParam
}

func genMonthArr(months []string) ([13]bool, error) {
	monthsArr := [13]bool{}
	if len(months) == 0 {
		return [13]bool{true,true,true,true,true,true,true,true,true,true,true,true,true}, nil
	}

	for _, month := range months {
		val, err := strconv.ParseInt(month, 10, 0)
		if err != nil {
			return monthsArr, ErrInvalidMonthParam
		}
		if val < 1 || val > 12 {
			return monthsArr, ErrInvalidMonthParam
		}
		monthsArr[val] = true
	}

	return monthsArr, nil 
}

func genDaysArr(daysOfMonth []string, date time.Time) ([32]bool, error) {
	days := [32]bool{}

	for _, strVal := range daysOfMonth {
		val, err := strconv.ParseInt(strVal, 10, 0)
		if err != nil {
			return days, ErrInvalidMonthParam
		}
		if val < -2 || val > 31 {
			return days, ErrInvalidMonthParam
		}
		switch val {
		case -1, -2:
			numDay := calcDayMonth(int(val), date)
			days[numDay] = true
		default:
			days[val] = true
		}
	}

	return days, nil
}

func NextDate(now time.Time, dstart, repeat string) (string, error) {
	// парсинг исходной даты
	date, err := time.Parse(constants.TimeFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга dstart: %w", err)
	}

	// парсинг правил повтора
	rawData := strings.Fields(repeat)
	if len(rawData) < 1 {
		return "", ErrEmptyRepeat
	}
	switch rawData[0] {
	case "d":
		if len(rawData) <= 1 {
			return "", ErrEmptyDay
		}
		days, err := strconv.ParseInt(rawData[1], 10, 0)
		if err != nil {
			return "", ErrInvalidFormat
		}
		if days < 1 || days > 400 {
			return "", ErrInvalidInterval
		}
		for {
			date = date.AddDate(0, 0, int(days))
			if afterNow(date, now) {
				break
			}
		}
	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}
	case "w":
		if len(rawData) <= 1 {
			return "", ErrInvalidFormat
		}

		daysOfWeek := strings.Split(rawData[1], ",")

		for {
			date = date.AddDate(0,0,1)
			ok, err := isDayWeek(date, daysOfWeek)
			if err != nil {
				return "", err
			}
			if ok && afterNow(date, now) {
				break
			}
		}
	case "m":
		if len(rawData) <= 1 {
			return "", ErrInvalidFormat
		}

		daysOfMonth := strings.Split(rawData[1], ",")
		// заполнение опционального параметра
		var months []string
		if len(rawData) == 3 {
			months = strings.Split(rawData[2], ",")
		}
		// генерация булевого массива месяцев
		monthArr, err := genMonthArr(months)
		if err != nil {
			return "", err
		}
		
		for {
			date = date.AddDate(0,0,1)
			var days [32]bool
			// если нет такого месяца, то следующая итерация
			if !monthArr[date.Month()] {
				continue
			}
			days, err = genDaysArr(daysOfMonth, date)
			if err != nil {
				return "", err
			}
			// если нет такого дня, то следующая итерация
			if !days[date.Day()] {
				continue
			}
			if afterNow(date, now) {
				break
			}
		}

	default:
		return "", ErrInvalidFormat
	}

	// Финальное пробразование даты в строку
	return date.Format(constants.TimeFormat), nil
}

// специфичная  функция для определения последнего и предпоследнего дня в месяце
// sub - должно передаваться с минусом "-1" или "-2"
func calcDayMonth(sub int, date time.Time) int {
	if date.Month() == 12 {
		return 32 - sub
	}
	firstDayOfNextMonth := time.Date(date.Year(), date.Month() + 1, 1, 0,0,0,0,date.Location())
	dayMonth := firstDayOfNextMonth.AddDate(0,0,sub)
	return dayMonth.Day()
}