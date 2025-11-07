package export

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
	"github.com/xuri/excelize/v2"
)

type XlsxExporter struct{}

func NewXLSX() *XlsxExporter {
	return &XlsxExporter{}
}

func (e XlsxExporter) Export(data []domain.EmplCompleteData, writer fyne.URIWriteCloser) error {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{
		"ID",
		"ФИО",
		"Email",
		"Телефон",
		"Отработано (ч)",
		"Поздние приходы (ч)",
		"Ранние уходы (ч)",
	}
	startDynamicCol := len(headers) + 1
	for _, d := range data[0].DailyMarks {
		headers = append(headers, d.Date)
	}

	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for rowIndex, item := range data {
		row := rowIndex + 2

		values := []any{
			item.ID,
			item.Name,
			item.Email,
			item.PhoneNum,
			fmt.Sprintf("%.2f", item.WorkedTime),
			fmt.Sprintf("%.2f", item.LateComeTime),
			fmt.Sprintf("%.2f", item.EarlyExitTime),
		}

		for colIndex, v := range values {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, row)
			f.SetCellValue(sheet, cell, v)
		}

		for j, d := range item.DailyMarks {
			col := startDynamicCol + j

			cell, _ := excelize.CoordinatesToCellName(col, row)

			f.SetCellValue(sheet, cell, d.WorkingTime+"\n"+d.ComingTime+"-"+d.ExitingTime)
		}
	}

	return f.Write(writer)
}
