package database

import "time"

type Employee struct {
	ID       string `gorm:"primaryKey"`
	FullName string
	Email    string
	Phone    string
}

type MonthlyData struct {
	ID             uint   `gorm:"primaryKey"`
	EmployeeID     string `gorm:"index"`
	Year           int
	Month          int
	WorkedHours    float64
	LateHours      float64
	EarlyExitHours float64

	// Уникальность по сотруднику + месяц
	UniqueKey string `gorm:"uniqueIndex"`
}

type DailyMark struct {
	ID         uint      `gorm:"primaryKey"`
	EmployeeID string    `gorm:"index"`
	Date       time.Time `gorm:"index"` // формат YYYY-MM-DD
	WorkHours  string
	ComeTime   string
	ExitTime   string

	// Уникальность по сотруднику + дате
	UniqueKey string `gorm:"uniqueIndex"`
}
