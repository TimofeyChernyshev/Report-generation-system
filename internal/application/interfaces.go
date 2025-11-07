package application

import (
	"fyne.io/fyne/v2"
	"github.com/TimofeyChernyshev/Report-generation-system/internal/domain"
)

type FileRepository interface {
	GetFilesInDirectory(dirPath string) ([]domain.FileInfo, error)
	LoadFile(filepath string) ([]domain.EmplRawData, error)
}

type Exporter interface {
	Export(data []domain.EmplCompleteData, writer fyne.URIWriteCloser) error
}
