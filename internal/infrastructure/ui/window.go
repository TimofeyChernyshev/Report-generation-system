package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

// Window управляет окнами приложения
type Window struct {
	app           fyne.App
	Window        fyne.Window
	reportService *application.ReportService

	selectedFolderLabel *widget.Label
	fileList            *widget.List
	currentFiles        []domain.FileInfo

	rawExamplesTable *widget.Table
	rawExamples      map[string][]domain.EmplRawData
	selectedFile     string
}

// NewWindowManager создает новый экземпляр Window
func NewWindow(app fyne.App, reportService *application.ReportService) *Window {
	w := &Window{app: app, reportService: reportService}
	w.Window = app.NewWindow("Report Generation System")
	w.Window.Resize(fyne.NewSize(800, 600))

	// Список файлов
	w.selectedFolderLabel = widget.NewLabel("Папка не выбрана")
	w.fileList = widget.NewList(
		func() int { return len(w.currentFiles) },
		func() fyne.CanvasObject { return widget.NewButton("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Button).SetText(w.currentFiles[i].Path)
			o.(*widget.Button).OnTapped = func() {
				w.handlePrewiewFile(w.currentFiles[i].Path)
			}
		},
	)

	// Таблица сырых данных о сотрудниках
	w.rawExamplesTable = widget.NewTable(
		func() (int, int) {
			return len(w.rawExamples[w.selectedFile]) + 1, 5
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				switch id.Col {
				case 0:
					label.SetText("ID")
				case 1:
					label.SetText("Рабочее время")
				case 2:
					label.SetText("Приход")
				case 3:
					label.SetText("Уход")
				case 4:
					label.SetText("ФИО")
				}
				return
			}
			data := w.rawExamples[w.selectedFile][id.Row-1]
			switch id.Col {
			case 0:
				label.SetText(data.ID)
			case 1:
				label.SetText(data.WorkingTime)
			case 2:
				label.SetText(data.ComingTime)
			case 3:
				label.SetText(data.ExitingTime)
			case 4:
				label.SetText(data.Name)
			}
		},
	)
	w.rawExamplesTable.SetColumnWidth(0, 80)
	w.rawExamplesTable.SetColumnWidth(1, 120)
	w.rawExamplesTable.SetColumnWidth(2, 100)
	w.rawExamplesTable.SetColumnWidth(3, 100)
	w.rawExamplesTable.SetColumnWidth(4, 200)

	w.rawExamplesTable.Hide()

	scroll := container.NewVScroll(w.rawExamplesTable)
	scroll.SetMinSize(fyne.NewSize(600, 10*25)) // 10 видимых строк

	// Кнопки в системе
	selectFolderBtn := widget.NewButton("Выбрать папку", w.handleSelectFolder)
	// selectFilePrewiew := widget.New

	content := container.NewVBox(
		container.NewHBox(selectFolderBtn, w.selectedFolderLabel),
		w.fileList,
		scroll,
	)

	w.Window.SetContent(content)
	return w
}
