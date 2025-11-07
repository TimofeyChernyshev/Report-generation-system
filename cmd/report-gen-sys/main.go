package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/export"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/files"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/ui"
)

func main() {
	a := app.NewWithID("1")

	fileRepo := files.NewFileRepository()

	exporters := make(map[string]application.Exporter)
	exporters[".csv"] = export.NewCSV()
	exporters[".xlsx"] = export.NewXLSX()

	reportService := application.NewReportService(fileRepo, exporters)
	window := ui.NewWindow(a, reportService)
	window.Window.Show()
	a.Run()
}
