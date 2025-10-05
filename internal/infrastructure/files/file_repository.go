package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

// FileRepositoryImpl реализация FileRepository
type FileRepositoryImpl struct{}

// NewFileRepository создает новую реализацию FileRepository
func NewFileRepository() domain.FileRepository {
	return &FileRepositoryImpl{}
}

// Вовзращает все json файлы из дирректории
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

func (r *FileRepositoryImpl) LoadJSONFile(path string) (*domain.FileInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("parsing JSON error: %w", err)
	}

	jsonFile := &domain.FileInfo{
		Path:    path,
		Content: content,
	}

	return jsonFile, nil
}
