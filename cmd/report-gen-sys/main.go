package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	database "github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/db"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/export"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/files"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/ui"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbSQLite, err := gorm.Open(sqlite.Open("employees.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	dbSQLite.AutoMigrate(
		&database.Employee{},
		&database.MonthlyData{},
		&database.DailyMark{},
	)

	a := app.NewWithID("1")

	fileRepo := files.NewFileRepository()

	exporters := make(map[string]application.Exporter)
	exporters[".csv"] = export.NewCSV()
	exporters[".xlsx"] = export.NewXLSX()

	reportService := application.NewReportService(fileRepo, exporters, dbSQLite)
	window := ui.NewWindow(a, reportService)
	window.Window.Show()
	a.Run()
}
