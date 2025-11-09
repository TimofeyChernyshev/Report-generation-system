package ui

import (
	"fmt"
	"slices"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

type ColumnType int

const (
	ColString ColumnType = iota
	ColFloat
)

var columnTypes = map[int]ColumnType{
	0: ColString, // ID
	1: ColString, // ФИО
	2: ColString, // Email
	3: ColString, // Телефон
	4: ColFloat,  // Отработано
	5: ColFloat,  // Опоздания
	6: ColFloat,  // Ранние уходы
}

func (w *Window) showFilterMenu(col int) {
	switch columnTypes[col] {
	case ColString:
		w.showStringFilter(col)
	case ColFloat:
		w.showFloatFilter(col)
	}
}

func findUniqueValues(slice []domain.EmplCompleteData, col int) map[string]struct{} {
	set := map[string]struct{}{}
	for _, row := range slice {
		switch col {
		case 0:
			set[row.ID] = struct{}{}
		case 1:
			set[row.Name] = struct{}{}
		case 2:
			set[row.Email] = struct{}{}
		case 3:
			set[row.PhoneNum] = struct{}{}
		default:
			set[row.DailyMarks[col-7].WorkingTime] = struct{}{}
		}
	}
	return set
}

// Фильтрация полей со значениями типа string
// Можно выбрать сотрудников по их ID, ФИО, имени или почте
func (w *Window) showStringFilter(col int) {
	valuesSet := findUniqueValues(w.dataSlice, col)
	filteredValueSet := findUniqueValues(w.filteredDataSlice, col)

	var values []string
	var filteredValues []string
	for v := range valuesSet {
		values = append(values, v)
	}
	for v := range filteredValueSet {
		filteredValues = append(filteredValues, v)
	}
	sort.Strings(values)

	checkboxes := []*widget.Check{}
	for _, v := range values {
		cb := widget.NewCheck(v, nil)
		if slices.Contains(filteredValues, v) {
			cb.SetChecked(true)
		} else {
			cb.SetChecked(false)
		}
		checkboxes = append(checkboxes, cb)
	}

	list := container.NewVBox()
	for _, cb := range checkboxes {
		list.Add(cb)
	}
	scrolled := container.NewVScroll(list)
	scrolled.SetMinSize(fyne.NewSize(300, 180))

	dialog.ShowCustomConfirm("Фильтр", "Применить", "Отмена", scrolled,
		func(apply bool) {
			if !apply {
				return
			}

			selected := map[string]struct{}{}
			for _, cb := range checkboxes {
				if cb.Checked {
					selected[cb.Text] = struct{}{}
				}
			}

			filtered := make([]domain.EmplCompleteData, 0)
			for _, row := range w.dataSlice {
				var val string
				switch col {
				case 0:
					val = row.ID
				case 1:
					val = row.Name
				case 2:
					val = row.Email
				case 3:
					val = row.PhoneNum
				default:
					val = row.DailyMarks[col-7].WorkingTime
				}
				if _, ok := selected[val]; ok {
					filtered = append(filtered, row)
				}
			}
			w.filteredDataSlice = filtered
			w.completeDataTable.Refresh()
		},
		w.Window)
}

// Фильтрация полей со значениями в виде float
// При использовании этой фильтрации учитываются все данные и отменяется фильтрация string
// Однако можно применить string фильтрацию поверх float фильтрации
func (w *Window) showFloatFilter(col int) {
	compareList := widget.NewSelect([]string{">", "<", "=", "≥", "≤", "≠"}, nil)
	compareList.SetSelected(">")

	valueInput := widget.NewEntry()
	valueInput.SetPlaceHolder("Число")

	dialog.ShowForm("Фильтр по значению", "Применить", "Отмена",
		[]*widget.FormItem{
			{Text: "Условие", Widget: compareList},
			{Text: "Значение", Widget: valueInput},
		},
		func(apply bool) {
			if !apply {
				return
			}
			if valueInput.Text == "" {
				return
			}

			f, err := strconv.ParseFloat(valueInput.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("неверное число: %s", valueInput.Text), w.Window)
				return
			}

			filtered := make([]domain.EmplCompleteData, 0)
			for _, row := range w.dataSlice {
				var val float64
				switch col {
				case 4:
					val = row.WorkedTime
				case 5:
					val = row.LateComeTime
				case 6:
					val = row.EarlyExitTime
				}
				if compareList.Selected == ">" && val > f {
					filtered = append(filtered, row)
				}
				if compareList.Selected == "<" && val < f {
					filtered = append(filtered, row)
				}
				if compareList.Selected == "=" && val == f {
					filtered = append(filtered, row)
				}
				if compareList.Selected == "≥" && val >= f {
					filtered = append(filtered, row)
				}
				if compareList.Selected == "≤" && val <= f {
					filtered = append(filtered, row)
				}
				if compareList.Selected == "≠" && val != f {
					filtered = append(filtered, row)
				}
			}
			w.filteredDataSlice = filtered
			w.completeDataTable.Refresh()
		},
		w.Window)
}
