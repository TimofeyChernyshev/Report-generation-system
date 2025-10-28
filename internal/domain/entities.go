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
	WorkedTime    float64 // Отработанное за месяц время в часах
	LateComeTime  float64 // Общее время опозданий в часах
	EarlyExitTime float64 // Общее время ранних уходов в часах
	DailyMarks    []Mark  // Ежедневные отметки по пользователю
}

type Mark struct {
	WorkingTime string
	ComingTime  string
	ExitingTime string
	Date        string
}
