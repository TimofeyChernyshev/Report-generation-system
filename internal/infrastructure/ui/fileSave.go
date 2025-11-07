package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	lastFolderKey = "fyne:fileDialogLastFolder"
)

type favoriteItem struct {
	locName string
	locIcon fyne.Resource
	loc     fyne.URI
}

type fileDialog struct {
	file             *FileDialog
	fileName         *widget.Entry
	title            *widget.Label
	dismiss          *widget.Button
	open             *widget.Button
	breadcrumb       *fyne.Container
	breadcrumbScroll *container.Scroll
	files            *widget.List
	filesScroll      *container.Scroll
	favorites        []favoriteItem
	favoritesList    *widget.List

	formatSelect *widget.Select

	data []fyne.URI

	win        *widget.PopUp
	selected   fyne.URI
	selectedID int
	dir        fyne.ListableURI
}

// FileDialog is a dialog containing a file picker for use in saving files.
type FileDialog struct {
	callback         func(fyne.URIWriteCloser, error)
	onClosedCallback func(bool)
	parent           fyne.Window
	dialog           *fileDialog

	titleText        string
	confirmText      string
	dismissText      string
	desiredSize      fyne.Size
	filter           storage.FileFilter
	startingLocation fyne.ListableURI
	SelectedFormat   string
}

func (f *fileDialog) makeUI() fyne.CanvasObject {
	saveName := widget.NewEntry()
	saveName.OnChanged = func(s string) {
		if s == "" {
			f.open.Disable()
		} else {
			f.open.Enable()
		}
	}
	saveName.SetPlaceHolder("Введите имя файла")
	saveName.OnSubmitted = func(s string) {
		f.open.OnTapped()
	}
	f.fileName = saveName

	label := "Сохранить"
	if f.file.confirmText != "" {
		label = f.file.confirmText
	}
	f.open = f.makeOpenButton(label)

	dismissLabel := "Отмена"
	if f.file.dismissText != "" {
		dismissLabel = f.file.dismissText
	}
	f.dismiss = f.makeDismissButton(dismissLabel)

	buttons := container.NewGridWithRows(1, f.dismiss, f.open)

	f.files = widget.NewList(
		func() int { return len(f.data) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.FileIcon()), widget.NewLabel("Template"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(f.data) {
				return
			}
			file := f.data[id]
			isDir, _ := storage.CanList(file)

			icon := theme.FileIcon()
			if isDir {
				icon = theme.FolderIcon()
			}

			item.(*fyne.Container).Objects[0].(*widget.Icon).SetResource(icon)
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(file.Name())
		},
	)
	f.files.OnSelected = func(id widget.ListItemID) {
		if id >= len(f.data) {
			return
		}
		file := f.data[id]
		if listable, _ := storage.CanList(file); listable {
			f.setLocation(file)
		} else {
			f.setSelected(file, id)
		}
	}

	f.filesScroll = container.NewScroll(f.files)
	f.filesScroll.SetMinSize(fyne.NewSize(400, 300))

	f.breadcrumb = container.NewHBox()
	f.breadcrumbScroll = container.NewHScroll(container.NewPadded(f.breadcrumb))

	title := "Save File"
	if f.file.titleText != "" {
		title = f.file.titleText
	}
	f.title = widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	f.loadFavorites()

	f.favoritesList = widget.NewList(
		func() int { return len(f.favorites) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(f.favorites) {
				return
			}
			item.(*fyne.Container).Objects[0].(*widget.Icon).SetResource(f.favorites[id].locIcon)
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(f.favorites[id].locName)
		},
	)
	f.favoritesList.OnSelected = func(id widget.ListItemID) {
		if id >= len(f.favorites) {
			return
		}
		f.setLocation(f.favorites[id].loc)
	}

	// Выбор формата файла
	f.formatSelect = widget.NewSelect([]string{"Excel (.xlsx)", "CSV (.csv)"}, nil)
	f.formatSelect.SetSelected("Excel (.xlsx)")
	f.file.SelectedFormat = f.formatSelect.Selected

	f.formatSelect.OnChanged = func(selected string) {
		f.file.SelectedFormat = selected
	}

	formatContainer := container.NewHBox(
		widget.NewLabel("Расширение:"),
		f.formatSelect,
	)

	header := container.NewBorder(nil, nil, nil, nil, f.title)

	footer := container.NewBorder(
		formatContainer,
		nil,
		nil,
		buttons,
		container.NewHScroll(f.fileName),
	)

	body := container.NewHSplit(
		f.favoritesList,
		container.NewBorder(f.breadcrumbScroll, nil, nil, nil, f.filesScroll),
	)
	body.SetOffset(0.2)

	return container.NewBorder(header, footer, nil, nil, body)
}

func (f *fileDialog) makeOpenButton(label string) *widget.Button {
	btn := widget.NewButton(label, func() {
		if f.file.callback == nil {
			f.win.Hide()
			if f.file.onClosedCallback != nil {
				f.file.onClosedCallback(false)
			}
			return
		}

		name := f.fileName.Text

		var extension string
		switch f.file.SelectedFormat {
		case "Excel (.xlsx)":
			extension = ".xlsx"
		case "CSV (.csv)":
			extension = ".csv"
		default:
			extension = ".xlsx"
		}
		name += extension

		location, _ := storage.Child(f.dir, name)
		exists, _ := storage.Exists(location)

		if !exists {
			f.win.Hide()
			if f.file.onClosedCallback != nil {
				f.file.onClosedCallback(true)
			}
			f.file.callback(storage.Writer(location))
			return
		}

		// Запрос подтверждения для перезаписи
		confirmDialog := dialog.NewConfirm("Перезаписать?",
			fmt.Sprintf("Вы уверены, что хотите перезаписать файл?\n%s", name),
			func(ok bool) {
				if !ok {
					return
				}
				f.win.Hide()
				f.file.callback(storage.Writer(location))
				if f.file.onClosedCallback != nil {
					f.file.onClosedCallback(true)
				}
			}, f.file.parent)
		confirmDialog.Show()
	})

	btn.Importance = widget.HighImportance
	btn.Disable()
	return btn
}

func (f *fileDialog) makeDismissButton(label string) *widget.Button {
	return widget.NewButton(label, func() {
		f.win.Hide()
		if f.file.onClosedCallback != nil {
			f.file.onClosedCallback(false)
		}
		if f.file.callback != nil {
			f.file.callback(nil, nil)
		}
	})
}

func (f *fileDialog) loadFavorites() {
	homeDir, _ := os.UserHomeDir()
	desktopDir := filepath.Join(homeDir, "Desktop")
	documentsDir := filepath.Join(homeDir, "Documents")
	downloadsDir := filepath.Join(homeDir, "Downloads")

	f.favorites = []favoriteItem{
		{locName: "Home", locIcon: theme.HomeIcon(), loc: storage.NewFileURI(homeDir)},
		{locName: "Desktop", locIcon: theme.DesktopIcon(), loc: storage.NewFileURI(desktopDir)},
		{locName: "Documents", locIcon: theme.DocumentIcon(), loc: storage.NewFileURI(documentsDir)},
		{locName: "Downloads", locIcon: theme.DownloadIcon(), loc: storage.NewFileURI(downloadsDir)},
	}
}

func (f *fileDialog) refreshDir(dir fyne.ListableURI) {
	f.data = nil

	files, err := dir.List()
	if err != nil {
		return
	}

	for _, file := range files {
		if isHidden(file) {
			continue
		}

		listable, err := storage.ListerForURI(file)
		if err == nil {
			f.data = append(f.data, listable)
		} else if f.file.filter == nil || f.file.filter.Matches(file) {
			f.data = append(f.data, file)
		}
	}

	// Сортируем: сначала папки, потом файлы
	sort.Slice(f.data, func(i, j int) bool {
		iDir, _ := storage.CanList(f.data[i])
		jDir, _ := storage.CanList(f.data[j])

		if iDir && !jDir {
			return true
		}
		if !iDir && jDir {
			return false
		}
		return strings.ToLower(f.data[i].Name()) < strings.ToLower(f.data[j].Name())
	})

	f.files.Refresh()
}

func (f *fileDialog) setLocation(dir fyne.URI) error {
	f.selectedID = -1
	if dir == nil {
		return errors.New("failed to open nil directory")
	}
	list, err := storage.ListerForURI(dir)
	if err != nil {
		return err
	}

	fyne.CurrentApp().Preferences().SetString(lastFolderKey, dir.String())
	f.setSelected(nil, -1)
	f.dir = list

	f.breadcrumb.Objects = nil
	localdir := dir.String()[len(dir.Scheme())+3:]

	buildDir := filepath.VolumeName(localdir)
	for i, d := range strings.Split(localdir, "/") {
		if d == "" {
			if i > 0 {
				break
			}
			buildDir = "/"
			d = "/"
		} else if i > 0 {
			buildDir = filepath.Join(buildDir, d)
		} else {
			d = buildDir
			buildDir = d + string(os.PathSeparator)
		}

		newDir := storage.NewFileURI(buildDir)
		f.breadcrumb.Add(
			widget.NewButton(d, func() {
				f.setLocation(newDir)
			}),
		)
	}

	f.breadcrumbScroll.Refresh()
	f.refreshDir(list)
	return nil
}

func (f *fileDialog) setSelected(file fyne.URI, id int) {
	f.selected = file
	f.selectedID = id

	if file == nil || file.String()[len(file.Scheme())+3:] == "" {
		f.fileName.SetText("")
		f.open.Disable()
	} else {
		fileName := file.Name()
		ext := filepath.Ext(fileName)
		fileNameNoExt := fileName[:len(fileName)-len(ext)]

		switch ext {
		case ".csv":
			f.formatSelect.SetSelected("CSV (.csv)")
		case ".xlsx":
			f.formatSelect.SetSelected("Excel (.xlsx)")
		default:
			f.formatSelect.SetSelected("Excel (.xlsx)")
		}
		f.fileName.SetText(fileNameNoExt)
		f.open.Enable()
	}
}

func (f *FileDialog) effectiveStartingDir() fyne.ListableURI {
	if f.startingLocation != nil {
		return f.startingLocation
	}

	lastPath := fyne.CurrentApp().Preferences().String(lastFolderKey)
	if lastPath != "" {
		parsed, err := storage.ParseURI(lastPath)
		if err == nil {
			dir, err := storage.ListerForURI(parsed)
			if err == nil {
				return dir
			}
		}
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		lister, err := storage.ListerForURI(storage.NewFileURI(homeDir))
		if err == nil {
			return lister
		}
	}

	lister, _ := storage.ListerForURI(storage.NewFileURI("/"))
	return lister
}

func showFileSave(file *FileDialog) *fileDialog {
	d := &fileDialog{file: file}
	ui := d.makeUI()

	d.win = widget.NewModalPopUp(ui, file.parent.Canvas())
	d.win.Resize(fyne.NewSize(600, 400))

	d.setLocation(file.effectiveStartingDir())
	d.win.Show()
	d.win.Canvas.Focus(d.fileName)
	return d
}

// NewFileSave creates a file dialog allowing the user to choose a file to save to.
func NewFileSave(callback func(writer fyne.URIWriteCloser, err error), parent fyne.Window) *FileDialog {
	return &FileDialog{callback: callback, parent: parent}
}

// Show shows the file dialog.
func (f *FileDialog) Show() {
	if f.dialog != nil {
		f.dialog.win.Show()
		return
	}
	f.dialog = showFileSave(f)
	if !f.desiredSize.IsZero() {
		f.Resize(f.desiredSize)
	}
}

func (f *FileDialog) Resize(size fyne.Size) {
	f.desiredSize = size
	if f.dialog == nil {
		return
	}
	f.dialog.win.Resize(size)
}

// SetOnClosed sets a callback function that is called when the dialog is closed.
func (f *FileDialog) SetOnClosed(closed func()) {
	f.onClosedCallback = func(response bool) {
		closed()
	}
}

func isHidden(file fyne.URI) bool {
	return strings.HasPrefix(file.Name(), ".")
}
