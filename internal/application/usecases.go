package application

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

var (
	errIncorrectFile error = errors.New("no correct data in provided files")
)

// ReportService представляет систему создания отчетов посещаемости
type ReportService struct {
	fileRepo FileRepository
}

// NewReportService создает новый экземпляр ReportService
func NewReportService(fileRepo FileRepository) *ReportService {
	return &ReportService{
		fileRepo: fileRepo,
	}
}

func (s *ReportService) GetJSONFilesFromFolder(folderPath string) ([]domain.FileInfo, error) {
	return s.fileRepo.GetFilesInDirectory(folderPath)
}

// ImportAndValidateFiles открывает все файлы и проверяет валидность записей в них
func (s *ReportService) ImportAndValidateFiles(files []domain.FileInfo) (map[string][]domain.EmplRawData, []error) {
	var errs []error
	// ключ - дата из названия файла, значение - все записи из файла
	var records = make(map[string][]domain.EmplRawData)

	for _, file := range files {
		data, err := s.fileRepo.LoadFile(file.Path)
		if err != nil {
			errs = append(errs, fmt.Errorf("reading file(%s) error: %w", file.Path, err))
			continue
		}

		fileBase := filepath.Base(file.Path)
		fileExt := filepath.Ext(file.Path)
		fileName := fileBase[:len(fileBase)-len(fileExt)]
		records[fileName] = data
	}
	if len(records) == 0 {
		errs = append(errs, errIncorrectFile)
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
func (s *ReportService) CalculateTime(rawData map[string][]domain.EmplRawData, selectedDates map[string]bool) map[string]domain.EmplCompleteData {
	var dates []string
	for date := range selectedDates {
		dates = append(dates, date)
	}
	sort.Strings(dates)

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

	for _, records := range rawData {
		for _, record := range records {
			employee := calculated[record.ID]

			for i, mark := range employee.DailyMarks {
				if mark.Date == record.Date {
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

			calculated[record.ID] = employee
		}
	}

	return calculated
}
