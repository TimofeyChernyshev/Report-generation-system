package ui

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// Обработчик кнопки выбора папки
func (w *Window) handleSelectFolder() {
	dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w.Window)
			return
		}
		if uri == nil {
			return
		}

		folderPath := uri.Path()
		w.selectedFolderLabel.SetText("Папка: " + folderPath)

		files, err := w.reportService.GetJSONFilesFromFolder(folderPath)
		if err != nil {
			dialog.ShowError(err, w.Window)
			return
		}

		w.currentFiles = files
		w.fileList.Refresh()

		examples, errors := w.reportService.ImportAndValidateFiles(files)
		if errors != nil {
			for _, e := range errors {
				dialog.ShowError(e, w.Window)
			}
			return
		}
		w.rawExamples = examples
	}, w.Window).Show()
}

func (w *Window) handlePrewiewFile(path string) {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	w.selectedFile = base[:len(base)-len(ext)]
	w.rawExamplesTable.Show()
	w.rawExamplesTable.Refresh()
}
