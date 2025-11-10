package application

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
	database "github.com/TimofeyChernyshev/Report-generation-system/internal/infrastructure/db"
	"gorm.io/gorm"
)

var (
	errIncorrectFile       error = errors.New("no correct data in provided files")
	errDateInWrongFormat   error = errors.New("date must match the format YYYY-MM-DD or DD.MM.YYYY")
	errFilesDifferentMonth error = errors.New("files with different month or years must contains in different folders")
)

// ReportService представляет систему создания отчетов посещаемости
type ReportService struct {
	fileRepo  FileRepository
	exporters map[string]Exporter
	db        *gorm.DB
}

// NewReportService создает новый экземпляр ReportService
func NewReportService(fileRepo FileRepository, exporters map[string]Exporter, db *gorm.DB) *ReportService {
	return &ReportService{
		fileRepo: fileRepo, exporters: exporters, db: db,
	}
}

func (s *ReportService) GetJSONFilesFromFolder(folderPath string) ([]domain.FileInfo, error) {
	return s.fileRepo.GetFilesInDirectory(folderPath)
}

// ImportAndValidateFiles открывает все файлы и проверяет валидность записей в них
func (s *ReportService) ImportAndValidateFiles(files []domain.FileInfo) (map[time.Time][]domain.EmplRawData, []error) {
	var errs []error
	// ключ - дата из названия файла, значение - все записи из файла
	var records = make(map[time.Time][]domain.EmplRawData)

	layouts := []string{"2006-01-02", "02.01.2006"}
	var monthAndYear time.Time

	for _, file := range files {
		data, err := s.fileRepo.LoadFile(file.Path)
		if err != nil {
			errs = append(errs, fmt.Errorf("reading file(%s) error: %w", file.Path, err))
			continue
		}

		fileBase := filepath.Base(file.Path)
		fileExt := filepath.Ext(file.Path)
		fileName := fileBase[:len(fileBase)-len(fileExt)]
		var d time.Time
		for _, l := range layouts {
			d, err = time.Parse(l, fileName)
			if err == nil {
				dY, dM, _ := d.Date()
				if monthAndYear.IsZero() {
					monthAndYear = time.Date(dY, dM, 1, 0, 0, 0, 0, time.UTC)
				}
				if monthAndYear.Year() != dY || monthAndYear.Month() != dM {
					errs = append(errs, errFilesDifferentMonth)
					return nil, errs
				}
				break
			}
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%w: %s", errDateInWrongFormat, fileName))
			continue
		}
		records[d] = data
	}
	if len(records) == 0 {
		errs = append(errs, errIncorrectFile)
		return nil, errs
	}
	if len(errs) != 0 {
		return nil, errs
	}

	for date, data := range records {
		records[date], errs = validateAndNormalizeEmplData(data)
		record := records[date]
		for i := range record {
			record[i].Date = date
		}
		records[date] = record

		if len(errs) != 0 {
			return nil, errs
		}
	}

	return records, nil
}

// CalculateTime проводит все расчеты по сотрудникам
func (s *ReportService) CalculateTime(rawData map[time.Time][]domain.EmplRawData, selectedDates map[time.Time]bool) map[string]domain.EmplCompleteData {
	var dates []time.Time
	for date := range selectedDates {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// ключ - ID сотрудника, значение - его рассчитанные данные
	calculated := make(map[string]domain.EmplCompleteData)
	for _, records := range rawData {
		for _, record := range records {
			if _, exists := calculated[record.ID]; !exists {
				employee := domain.NewEmplCompleteData(record)
				employee.DailyMarks = make([]domain.Mark, len(dates))
				for i, date := range dates {
					employee.DailyMarks[i] = domain.Mark{
						WorkingTime: "",
						ComingTime:  "",
						ExitingTime: "",
						Date:        date,
					}
				}
				calculated[record.ID] = employee
			}
		}
	}

	for date, records := range rawData {
		for _, record := range records {
			employee := calculated[record.ID]

			for i, mark := range employee.DailyMarks {
				yM, mM, dM := mark.Date.Date()
				yR, mR, dR := record.Date.Date()
				if yM == yR && mM == mR && dM == dR {
					employee.DailyMarks[i] = domain.Mark{
						WorkingTime: record.WorkingTime,
						ComingTime:  record.ComingTime,
						ExitingTime: record.ExitingTime,
						Date:        record.Date,
					}
				}
			}

			if record.ComingTime == "" && record.ExitingTime != "" {
				record.ComingTime = record.WorkingTime[:5]
			}
			if record.ExitingTime == "" && record.ComingTime != "" {
				record.ExitingTime = record.WorkingTime[6:]
			}

			employee.WorkedTime += record.CalculateWorkedTime()
			employee.LateComeTime += record.CalculateLateComeTime()
			employee.EarlyExitTime += record.CalculateEarlyExitTime()
			employee.YearAndMonth = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

			calculated[record.ID] = employee
		}
	}

	return calculated
}

func (s *ReportService) Export(ext string, data []domain.EmplCompleteData, writer fyne.URIWriteCloser) error {
	exporter, ok := s.exporters[ext]
	if !ok {
		return fmt.Errorf("format %s unsupported", ext)
	}
	return exporter.Export(data, writer)
}

func (s *ReportService) SaveReportResults(data []domain.EmplCompleteData) {
	year, mon := data[0].YearAndMonth.Year(), int(data[0].YearAndMonth.Month())

	for _, emp := range data {
		// Сохранить / обновить сотрудника
		employee := database.Employee{
			ID:       emp.ID,
			FullName: emp.Name,
			Email:    emp.Email,
			Phone:    emp.PhoneNum,
		}
		s.db.Where("id = ?", emp.ID).FirstOrCreate(&employee)

		// Сохранить / обновить месяц
		monthModel := database.MonthlyData{
			EmployeeID:     employee.ID,
			Year:           year,
			Month:          mon,
			WorkedHours:    emp.WorkedTime,
			LateHours:      emp.LateComeTime,
			EarlyExitHours: emp.EarlyExitTime,
		}

		var existingMonth database.MonthlyData
		s.db.Where("employee_id = ? AND year = ? AND month = ?",
			employee.ID, year, mon).First(&existingMonth)

		if existingMonth.ID == 0 {
			s.db.Create(&monthModel)
		} else {
			s.db.Model(&existingMonth).Updates(monthModel)
		}

		// Сохранить / обновить ежедневные отметки
		for _, d := range emp.DailyMarks {
			daily := database.DailyMark{
				EmployeeID: employee.ID,
				Date:       d.Date,
				WorkHours:  d.WorkingTime,
				ComeTime:   d.ComingTime,
				ExitTime:   d.ExitingTime,
			}

			var existingDaily database.DailyMark
			s.db.Where("employee_id = ? AND date = ?", employee.ID, d.Date).First(&existingDaily)

			if existingDaily.ID == 0 {
				s.db.Create(&daily)
			} else {
				s.db.Model(&existingDaily).Updates(daily)
			}
		}
	}

}
