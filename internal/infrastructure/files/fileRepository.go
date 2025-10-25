package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TimofeyChernyshev/Report-generation-system/internal/application"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

// FileRepositoryImpl реализация FileRepository
type FileRepositoryImpl struct{}

// NewFileRepository создает новую реализацию FileRepository
func NewFileRepository() application.FileRepository {
	return &FileRepositoryImpl{}
}

// GetFilesInDirectory возвращает все json файлы из дирректории
func (r *FileRepositoryImpl) GetFilesInDirectory(path string) ([]domain.FileInfo, error) {
	var files []domain.FileInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			fullPath := filepath.Join(path, entry.Name())
			files = append(files, domain.FileInfo{
				Path: fullPath,
			})
		}
	}

	return files, nil
}

// LoadJSONFile парсит данные из json файла
func (r *FileRepositoryImpl) LoadFile(path string) ([]domain.EmplRawData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var temp []domain.EmplRawData
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("parsing JSON error: %w", err)
	}

	employees := make([]domain.EmplRawData, len(temp))
	for i, d := range temp {
		employees[i] = domain.EmplRawData{
			ID:          d.ID,
			Name:        d.Name,
			Email:       d.Email,
			PhoneNum:    d.PhoneNum,
			WorkingTime: d.WorkingTime,
			ComingTime:  d.ComingTime,
			ExitingTime: d.ExitingTime,
		}
	}

	return employees, nil
}
