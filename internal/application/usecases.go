package application

import (
	"errors"
	"fmt"
	"path/filepath"

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
		if len(errs) != 0 {
			return nil, errs
		}
	}

	return records, nil
}
