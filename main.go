package main

import (
	"io/ioutil"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type config struct {
	EditWidget    *widget.Entry
	PreviewWidget *widget.RichText
	CurrentFile   fyne.URI
	SaveMenuItem  *fyne.MenuItem
}

var cfg config

func main() {
	// create a fyne app
	a := app.New()
	a.Settings().Theme().Icon("icon.png")

	// create a windows for the app
	w := a.NewWindow("Markdown")

	// get the user interface
	edit, preview := cfg.makeUI()
	cfg.createMenuItem(w)

	// set the content of the window
	w.SetContent(container.NewHSplit(edit, preview))

	// show windows and run
	w.Resize(fyne.Size{Width: 800, Height: 500})
	w.CenterOnScreen()
	w.ShowAndRun()
}

func (app *config) makeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")
	app.EditWidget = edit
	app.PreviewWidget = preview

	edit.OnChanged = preview.ParseMarkdown

	return edit, preview
}

func (app *config) createMenuItem(w fyne.Window) {
	// create three menu items

	openMenuItem := fyne.NewMenuItem("Open...", app.openFunc(w))

	saveMenuItem := fyne.NewMenuItem("Save...", app.saveFunc(w))

	app.SaveMenuItem = saveMenuItem
	app.SaveMenuItem.Disabled = false

	saveAsMenuItem := fyne.NewMenuItem("Save as...", app.saveAsFunc(w))

	// create a file menu, and add the three items to it
	fileMenu := fyne.NewMenu("File", openMenuItem, saveMenuItem, saveAsMenuItem)

	menu := fyne.NewMainMenu(fileMenu)

	// set main menu
	w.SetMainMenu(menu)
}

var filter = storage.NewExtensionFileFilter([]string{".md", ".MD"})

func (app *config) openFunc(w fyne.Window) func() {
	return func() {
		openDialog := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			if read == nil {
				return
			}
			defer read.Close()

			data, err := ioutil.ReadAll(read)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			app.EditWidget.SetText(string(data))

			app.CurrentFile = read.URI()
			w.SetTitle(w.Title() + " - " + read.URI().Name())
			app.SaveMenuItem.Disabled = false

		}, w)
		openDialog.SetFilter(filter)
		openDialog.Show()
	}
}

func (app *config) saveFunc(w fyne.Window) func() {
	return func() {
		if app.CurrentFile != nil {
			write, err := storage.Writer(app.CurrentFile)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			write.Write([]byte(app.EditWidget.Text))
			defer write.Close()

		}
	}
}

func (app *config) saveAsFunc(w fyne.Window) func() {
	return func() {
		saveDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if write == nil {
				// user cancelled
				return
			}
			if !strings.HasSuffix(strings.ToLower(write.URI().String()), ".md") {
				dialog.ShowInformation("Error", "Please name your file with a .md extension!!", w)
				return
			}

			// save file
			write.Write([]byte(app.EditWidget.Text))
			app.CurrentFile = write.URI()

			defer write.Close()
			w.SetTitle(w.Title() + " - " + write.URI().Name())
			app.SaveMenuItem.Disabled = false

		}, w)
		saveDialog.SetFileName("untitled.md")
		saveDialog.SetFilter(filter)
		saveDialog.Show()
	}
}
