package domain

type FileRepository interface {
	GetFilesInDirectory(dirPath string) ([]FileInfo, error)
	LoadJSONFile(filepath string) (*FileInfo, error)
}
