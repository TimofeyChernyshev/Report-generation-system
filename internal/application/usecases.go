package application

import (
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

type ReportService struct {
	fileRepo domain.FileRepository
}

// NewReportService создает новый экземпляр ReportService
func NewReportService(fileRepo domain.FileRepository) *ReportService {
	return &ReportService{
		fileRepo: fileRepo,
	}
}

func (s *ReportService) GetJSONFilesFromFolder(folderPath string) ([]domain.FileInfo, error) {
	return s.fileRepo.GetFilesInDirectory(folderPath)
}
