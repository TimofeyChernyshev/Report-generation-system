package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (w *Window) handleSelectFolder() {
	dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}
		if uri == nil {
			return
		}

		folderPath := uri.Path()
		w.selectedFolderLabel.SetText("Папка: " + folderPath)

		files, err := w.reportService.GetJSONFilesFromFolder(folderPath)
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}

		w.currentFiles = files
		w.fileList.Refresh()
	}, w.window).Show()
}
