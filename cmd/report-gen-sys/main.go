package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/files"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/ui"
)

func main() {
	a := app.NewWithID("1")

	fileRepo := files.NewFileRepository()
	uiManager := ui.NewWindow(a)

	reportService := application.NewReportService(fileRepo)

	uiManager.SetReportService(reportService)
	a.Run()
}
