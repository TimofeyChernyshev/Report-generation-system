package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

// WindowManager управляет окнами приложения
type Window struct {
	app           fyne.App
	window        fyne.Window
	reportService *application.ReportService

	selectedFolderLabel *widget.Label
	fileList            *widget.List
	currentFiles        []domain.FileInfo
}

// NewWindowManager создает новый WindowManager
func NewWindow(app fyne.App) *Window {
	w := &Window{app: app}
	w.window = app.NewWindow("Report Generation System")
	w.window.Resize(fyne.NewSize(800, 600))

	// Список файлов
	w.selectedFolderLabel = widget.NewLabel("Папка не выбрана")
	w.fileList = widget.NewList(
		func() int { return len(w.currentFiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(w.currentFiles[i].Path)
		},
	)

	// Кнопки в системе
	selectFolderBtn := widget.NewButton("Выбрать папку", w.handleSelectFolder)

	content := container.NewVBox(
		container.NewHBox(selectFolderBtn, w.selectedFolderLabel),
		w.fileList,
	)

	w.window.SetContent(content)
	w.window.Show()
	return w
}

// SetReportService устанавливает сервис отчетов
func (wm *Window) SetReportService(service *application.ReportService) {
	wm.reportService = service
}
