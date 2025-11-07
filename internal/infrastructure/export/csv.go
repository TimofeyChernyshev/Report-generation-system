package export

import (
	"encoding/csv"
	"strconv"

	"fyne.io/fyne/v2"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

type CSVExporter struct{}

func NewCSV() *CSVExporter {
	return &CSVExporter{}
}

func (e CSVExporter) Export(data []domain.EmplCompleteData, writer fyne.URIWriteCloser) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	row := []string{"ID", "ФИО", "Email", "Телефон", "Отработано (ч)", "Поздние приходы (ч)", "Ранние уходы (ч)"}
	for _, d := range data[0].DailyMarks {
		row = append(row, d.Date)
	}
	if err := w.Write(row); err != nil {
		return err
	}

	for _, item := range data {
		row := []string{item.ID, item.Name, item.Email, item.PhoneNum,
			strconv.FormatFloat(item.WorkedTime, 'f', 2, 64),
			strconv.FormatFloat(item.LateComeTime, 'f', 2, 64),
			strconv.FormatFloat(item.EarlyExitTime, 'f', 2, 64)}
		for _, d := range item.DailyMarks {
			row = append(row, d.WorkingTime+": "+d.ComingTime+"-"+d.ExitingTime)
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}
