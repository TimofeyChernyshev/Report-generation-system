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

	reportService := application.NewReportService(fileRepo)
	window := ui.NewWindow(a, reportService)
	window.Window.Show()
	a.Run()
}
