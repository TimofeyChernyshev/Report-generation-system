package ui

import (
	"errors"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Обработчик кнопки выбора папки
func (w *Window) handleSelectFolder() {
	w.calculateTimeBtn.Hide()
	w.rawDataTable.Hide()
	w.completeDataTable.Hide()
	w.disclaimer.Hide()
	w.fileList.Hide()
	w.exportBtn.Hide()

	dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w.Window)
			return
		}
		if uri == nil {
			return
		}

		folderPath := uri.Path()
		w.selectedFolderLabel.SetText("Папка: " + folderPath)

		files, err := w.reportService.GetJSONFilesFromFolder(folderPath)
		if err != nil {
			dialog.ShowError(err, w.Window)
			return
		}

		w.currentFiles = files
		w.fileList.Refresh()

		examples, errors := w.reportService.ImportAndValidateFiles(files)
		if errors != nil {
			for _, e := range errors {
				dialog.ShowError(e, w.Window)
			}
			w.calculateTimeBtn.Hide()
			return
		}
		w.rawData = examples
		w.calculateTimeBtn.Show()
		w.disclaimer.Show()
		w.fileList.Show()
	}, w.Window).Show()
}

// Обработчик предпросмотра файлов
func (w *Window) handlePrewiewFile(path string) {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	w.selectedFile = base[:len(base)-len(ext)]
	w.rawDataTable.Show()
	w.rawDataTable.Refresh()
	w.showRawDataTable()
}

// Обработчик кнопки расчета рабочего времени
func (w *Window) handleCalculateTime() {
	w.fileList.Hide()
	w.rawDataTable.Hide()
	w.calculateTimeBtn.Hide()
	w.disclaimer.Hide()

	var d dialog.Dialog

	selectedDates := make(map[string]bool)

	// Календарь для выбора дат
	currentMonth := time.Now()
	calendar := container.NewGridWithColumns(7)

	// Навигация по месяцам
	monthLabel := widget.NewLabelWithStyle(
		currentMonth.Format("January 2006"),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	updateCalendar := func(month time.Time) {
		currentMonth = month
		w.createCalendar(calendar, month, selectedDates)
		monthLabel.SetText(month.Format("January 2006"))
	}
	updateCalendar(currentMonth)

	prevMonthBtn := widget.NewButton("←", func() {
		updateCalendar(currentMonth.AddDate(0, -1, 0))
	})

	nextMonthBtn := widget.NewButton("→", func() {
		updateCalendar(currentMonth.AddDate(0, 1, 0))
	})

	navigation := container.NewHBox(
		prevMonthBtn,
		container.NewCenter(monthLabel),
		nextMonthBtn,
	)

	// Кнопка очистки выбора
	clearSelectionBtn := widget.NewButton("Очистить выбор", func() {
		selectedDates = make(map[string]bool)
		updateCalendar(currentMonth)
	})

	// Кнопка подтверждения
	confirmBtn := widget.NewButton("Продолжить", func() {
		w.completeData = w.reportService.CalculateTime(w.rawData, selectedDates)
		w.dataSlice = nil
		for _, data := range w.completeData {
			w.dataSlice = append(w.dataSlice, data)
		}
		sort.Slice(w.dataSlice, func(i, j int) bool {
			return w.dataSlice[i].Name < w.dataSlice[j].Name
		})
		w.showCompleteDataTable()
		w.exportBtn.Show()
		d.Dismiss()
	})

	// Кнопка отмены
	cancelBtn := widget.NewButton("Отмена", func() {
		w.fileList.Show()
		w.calculateTimeBtn.Show()
		w.disclaimer.Show()
		d.Dismiss()
	})

	buttons := container.NewHBox(layout.NewSpacer(), clearSelectionBtn, cancelBtn, confirmBtn)

	content := container.NewVBox(
		widget.NewLabel("Выберите отметки посещаемости, которые должны быть в отчете:"),
		navigation,
		calendar,
		buttons,
	)
	d = dialog.NewCustomWithoutButtons("Выбор дат", content, w.Window)
	d.Show()
}

// Создает новый календарь для месяца
func (w *Window) createCalendar(grid *fyne.Container, month time.Time, selectedDates map[string]bool) {
	grid.Objects = nil

	// Заголовки дней недели
	days := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	for _, day := range days {
		header := widget.NewLabelWithStyle(day, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		grid.Add(header)
	}

	// Заполняем календарь
	firstDay := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	// Пустые ячейки до первого дня
	weekday := int(firstDay.Weekday())
	if weekday == 0 { // Воскресенье
		weekday = 7
	}
	for i := 1; i < weekday; i++ {
		grid.Add(widget.NewLabel(""))
	}

	// Дни месяца
	for d := firstDay; d.Compare(lastDay) <= 0; d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("02.01.2006")

		dayBtn := widget.NewButton(strconv.Itoa(d.Day()), nil)
		if selectedDates[dateStr] {
			dayBtn.Importance = widget.HighImportance
		}

		dayBtn.OnTapped = func() {
			if selectedDates[dateStr] {
				delete(selectedDates, dateStr)
				dayBtn.Importance = widget.MediumImportance
			} else {
				selectedDates[dateStr] = true
				dayBtn.Importance = widget.HighImportance
			}
			dayBtn.Refresh()
		}

		grid.Add(dayBtn)
	}
	grid.Refresh()
}

func (w *Window) handleExport() {
	dialogSave := NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w.Window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		fileURI := writer.URI()
		if fileURI == nil {
			dialog.ShowError(errors.New("cannot get file URI"), w.Window)
			return
		}
		fileName := fileURI.Name()
		fileExt := strings.ToLower(filepath.Ext(fileName))

		exportErr := w.reportService.Export(fileExt, w.dataSlice, writer)
		if exportErr != nil {
			dialog.ShowError(exportErr, w.Window)
			return
		}
	}, w.Window)

	dialogSave.Show()
}
