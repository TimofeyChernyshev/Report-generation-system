package application

import "github.com/TimofeyChernyshev/Report-generation-system/internal/domain"

type FileRepository interface {
	GetFilesInDirectory(dirPath string) ([]domain.FileInfo, error)
	LoadFile(filepath string) ([]domain.EmplRawData, error)
}
