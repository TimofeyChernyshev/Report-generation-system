package ui

import (
	"sort"
	"strconv"
	"time"

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

	rawDataTable *widget.Table
	rawData      map[time.Time][]domain.EmplRawData
	selectedFile string

	calculateTimeBtn *widget.Button
	disclaimer       *widget.Label

	completeDataTable *widget.Table
	completeData      map[string]domain.EmplCompleteData
	dataSlice         []domain.EmplCompleteData
	filteredDataSlice []domain.EmplCompleteData

	exportBtn *widget.Button
}

// NewWindowManager создает новый экземпляр Window
func NewWindow(app fyne.App, reportService *application.ReportService) *Window {
	w := &Window{app: app, reportService: reportService}
	w.Window = app.NewWindow("Report Generation System")
	w.Window.Resize(fyne.NewSize(800, 600))

	w.selectedFolderLabel = widget.NewLabel("Папка не выбрана")
	w.createFileList()

	// Таблицы данных
	w.createRawDataTable()
	scrollRaw := container.NewVScroll(w.rawDataTable)
	scrollRaw.SetMinSize(fyne.NewSize(600, 10*25))

	w.createCompleteDataTable()
	scrollComplete := container.NewVScroll(w.completeDataTable)
	scrollComplete.SetMinSize(fyne.NewSize(600, 10*25))

	tableContainer := container.NewStack(scrollRaw)

	// Кнопки в системе
	selectFolderBtn := widget.NewButton("Выбрать папку", w.handleSelectFolder)
	w.calculateTimeBtn = widget.NewButton("Рассчитать время", w.handleCalculateTime)
	w.calculateTimeBtn.Hide()
	w.exportBtn = widget.NewButton("Экспортировать", w.handleExport)
	w.exportBtn.Hide()

	w.disclaimer = widget.NewLabel("Если отсутствует время прихода или ухода, то будет использовано время из требуемого рабочего диапазона")
	w.disclaimer.Hide()

	content := container.NewVBox(
		container.NewHBox(selectFolderBtn, w.selectedFolderLabel, w.calculateTimeBtn),
		w.disclaimer,
		w.fileList,
		tableContainer,
		w.exportBtn,
	)

	w.Window.SetContent(content)
	return w
}

// Список всех файлов из папки
func (w *Window) createFileList() {
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
}

func (w *Window) parseFileName() time.Time {
	layouts := []string{"2006-01-02", "02.01.2006"}
	var fileDate time.Time
	var err error
	for _, l := range layouts {
		fileDate, err = time.Parse(l, w.selectedFile)
		if err == nil {
			break
		}
	}
	return fileDate
}

// Таблица сырых данных о сотрудниках
func (w *Window) createRawDataTable() {
	w.rawDataTable = widget.NewTable(
		func() (int, int) {
			date := w.parseFileName()
			return len(w.rawData[date]) + 1, 7
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			date := w.parseFileName()
			sort.Slice(w.rawData[date], func(i, j int) bool {
				return w.rawData[date][i].Name < w.rawData[date][j].Name
			})
			label := cell.(*widget.Label)
			if id.Row == 0 {
				switch id.Col {
				case 0:
					label.SetText("ID")
				case 1:
					label.SetText("ФИО")
				case 2:
					label.SetText("Рабочее время")
				case 3:
					label.SetText("Приход")
				case 4:
					label.SetText("Уход")
				case 5:
					label.SetText("Почта")
				case 6:
					label.SetText("Телефон")
				}
				return
			}
			data := w.rawData[date][id.Row-1]
			switch id.Col {
			case 0:
				label.SetText(data.ID)
			case 1:
				label.SetText(data.Name)
			case 2:
				label.SetText(data.WorkingTime)
			case 3:
				label.SetText(data.ComingTime)
			case 4:
				label.SetText(data.ExitingTime)
			case 5:
				label.SetText(data.Email)
			case 6:
				label.SetText(data.PhoneNum)
			}
		},
	)
	w.rawDataTable.SetColumnWidth(0, 80)
	w.rawDataTable.SetColumnWidth(1, 200)
	w.rawDataTable.SetColumnWidth(2, 120)
	w.rawDataTable.SetColumnWidth(3, 100)
	w.rawDataTable.SetColumnWidth(4, 100)
	w.rawDataTable.SetColumnWidth(5, 200)
	w.rawDataTable.SetColumnWidth(6, 120)

	w.rawDataTable.Hide()
}

// Таблица сырых данных о сотрудниках
func (w *Window) createCompleteDataTable() {
	w.completeDataTable = widget.NewTable(
		func() (int, int) {
			if w.filteredDataSlice == nil {
				return 1, 7
			}
			return len(w.filteredDataSlice) + 1, 7 + len(w.dataSlice[0].DailyMarks)
		},
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			btn := widget.NewButton("▼", nil)
			btn.Hide()
			return container.NewBorder(nil, nil, nil, btn, lbl)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			c := cell.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			button := c.Objects[1].(*widget.Button)
			button.OnTapped = func() {
				w.showFilterMenu(id.Col)
			}
			if id.Row == 0 {
				switch id.Col {
				case 0:
					label.SetText("ID")
					button.Show()
				case 1:
					label.SetText("ФИО")
					button.Show()
				case 2:
					label.SetText("Почта")
					button.Show()
				case 3:
					label.SetText("Телефон")
					button.Show()
				case 4:
					label.SetText("Отработанное время, ч")
					button.Show()
				case 5:
					label.SetText("Опоздания, ч")
					button.Show()
				case 6:
					label.SetText("Ранние уходы, ч")
					button.Show()
				default:
					label.SetText((w.filteredDataSlice[id.Row].DailyMarks[id.Col-7].Date).Format("02.01.2006"))
				}
				return
			}

			button.Hide()
			data := w.filteredDataSlice[id.Row-1]
			w.completeDataTable.SetRowHeight(id.Row, 50)
			switch id.Col {
			case 0:
				label.SetText(data.ID)
				w.completeDataTable.SetColumnWidth(id.Col, 80)
			case 1:
				label.SetText(data.Name)
				w.completeDataTable.SetColumnWidth(id.Col, 200)
			case 2:
				label.SetText(data.Email)
				w.completeDataTable.SetColumnWidth(id.Col, 200)
			case 3:
				label.SetText(data.PhoneNum)
				w.completeDataTable.SetColumnWidth(id.Col, 120)
			case 4:
				label.SetText(strconv.FormatFloat(data.WorkedTime, 'f', 2, 64))
				w.completeDataTable.SetColumnWidth(id.Col, 210)
			case 5:
				label.SetText(strconv.FormatFloat(data.LateComeTime, 'f', 2, 64))
				w.completeDataTable.SetColumnWidth(id.Col, 150)
			case 6:
				label.SetText(strconv.FormatFloat(data.EarlyExitTime, 'f', 2, 64))
				w.completeDataTable.SetColumnWidth(id.Col, 160)
			default:
				mark := data.DailyMarks[id.Col-7]
				if mark.WorkingTime == "" && mark.ComingTime == "" && mark.ExitingTime == "" {
					label.SetText("Нет данных")
				} else {
					label.SetText(mark.WorkingTime + "\n" + mark.ComingTime + "-" + mark.ExitingTime)
				}
				w.completeDataTable.SetColumnWidth(id.Col, 120)
			}
		},
	)

	w.completeDataTable.Hide()
}

// Функция для показа сырых данных
func (w *Window) showRawDataTable() {
	content := w.Window.Content().(*fyne.Container)
	tableContainer := content.Objects[3].(*fyne.Container)

	tableContainer.Objects = nil
	scrollRaw := container.NewVScroll(w.rawDataTable)
	scrollRaw.SetMinSize(fyne.NewSize(600, 10*25))
	tableContainer.Add(scrollRaw)

	w.rawDataTable.Show()
	w.completeDataTable.Hide()
}

// Функция для показа обработанных данных
func (w *Window) showCompleteDataTable() {
	content := w.Window.Content().(*fyne.Container)
	tableContainer := content.Objects[3].(*fyne.Container)

	tableContainer.Objects = nil
	scrollComplete := container.NewVScroll(w.completeDataTable)
	scrollComplete.SetMinSize(fyne.NewSize(600, 10*25))
	tableContainer.Add(scrollComplete)

	w.rawDataTable.Hide()
	w.completeDataTable.Show()
}
