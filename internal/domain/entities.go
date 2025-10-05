package domain

type FileInfo struct {
	Path    string                 `json:"info"`
	Content map[string]interface{} `json:"content"`
}

type EmplRawData struct {
	ID          string
	Name        string // ФИО
	Email       string
	PhoneNum    string
	WorkingTime string // Требуемое рабочее время в формате HH:MM-HH:MM
	ComingTime  string // Время прихода на работу HH:MM
	ExitingTime string // Время ухода с работы HH:MM
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
