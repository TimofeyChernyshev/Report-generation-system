package domain

type FileInfo struct {
	Path string `json:"path"`
}

type EmplRawData struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"` // ФИО
	Email       string `json:"Email"`
	PhoneNum    string `json:"PhoneNum"`
	WorkingTime string `json:"WorkingTime"` // Требуемое рабочее время в формате HH:MM-HH:MM
	ComingTime  string `json:"ComingTime"`  // Время прихода на работу HH:MM
	ExitingTime string `json:"ExitingTime"` // Время ухода с работы HH:MM
	Date        string `json:"-"`           // Дата записи
}

type EmplCompleteData struct {
	ID            string
	Name          string
	Email         string
	PhoneNum      string
	WorkedTime    string // Отработанное за месяц время
	LateComeTime  string // Общее время опозданий
	EarlyExitTime string // Общее время ранних уходов
}
