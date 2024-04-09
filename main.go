package main

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// in is the currently used input for either encryption or decryption.
// It is kept as global state, so that it can never be lost to the GC
// before closing.
var in fyne.URIReadCloser

var windowSize = fyne.Size{Width: 800, Height: 480}
var dialogSize = fyne.Size{Width: 780, Height: 460}

const encryptOptionExplanation = `Use encryption to protect files
from unauthorized access by
locking them with a password.`

const decryptOptionExplanation = `Use decryption to regain access
to previously encrypted files by
providing the password.`

func main() {
	a := app.New()
	a.Settings().SetTheme(&myTheme{})
	w := a.NewWindow("spind")
	w.Resize(windowSize)
	if fillInFromArgs() {
		if strings.HasSuffix(strings.ToLower(os.Args[1]), ".age") {
			drawDecryptForm(w)
		} else {
			drawEncryptForm(w)
		}
	} else {
		drawMenu(w)
	}
	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) { handleFileDrop(w, uris) })
	w.ShowAndRun()
}

// fillInFromArgs fills the in variable with the file passed as
// os.Args[1], if present. If successful, true is returned.
func fillInFromArgs() bool {
	if len(os.Args) != 2 {
		return false
	}
	uri := storage.NewFileURI(os.Args[1])
	if droppedIn, err := storage.OpenFileFromURI(uri); err == nil {
		in = droppedIn
		return true
	}
	return false
}

func drawMenu(w fyne.Window) {
	if in != nil {
		// Coming from a sub-menu where an input file was selected.
		in.Close()
		in = nil
	}
	aboutButton := widget.NewToolbarAction(theme.QuestionIcon(), func() { showAbout(w) })
	top := widget.NewToolbar(widget.NewToolbarSpacer(), aboutButton)
	optionsBox := container.NewHBox(encryptOption(w), widget.NewSeparator(), decryptOption(w))
	options := container.NewPadded(container.NewCenter(optionsBox))
	hint := widget.NewLabel("Tip: You can drag and drop files into spind.")
	w.SetContent(container.New(layout.NewBorderLayout(top, hint, nil, nil), top, hint, options))
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleMenuKey(e, w) })
}

func handleMenuKey(e *fyne.KeyEvent, w fyne.Window) {
	switch e.Name {
	case fyne.KeyE:
		drawEncryptForm(w)
	case fyne.KeyD:
		drawDecryptForm(w)
	case fyne.KeyH, fyne.KeyF1:
		showAbout(w)
	case fyne.KeyQ:
		w.Close()
	}
}

func showAbout(w fyne.Window) {
	body := widget.NewRichTextFromMarkdown(`spind allows you to en- and decrypt files with passwords.

It is open source software and it's source code can be found at [github.com/codesoap/spind](https://github.com/codesoap/spind).

It uses the [age](https://age-encryption.org) file format and is compatible with other software using this format.`)
	d := dialog.NewCustom("About spind", "OK", body, w)
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, d) })
	d.SetOnClosed(func() {
		w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleMenuKey(e, w) })
	})
	d.Show()
}

func encryptOption(w fyne.Window) fyne.CanvasObject {
	text := widget.TextSegment{Style: widget.RichTextStyle{}, Text: encryptOptionExplanation}
	return container.NewVBox(
		container.NewCenter(widget.NewButton("Encrypt", func() { drawEncryptForm(w) })),
		widget.NewRichText(&text),
	)
}

func decryptOption(w fyne.Window) fyne.CanvasObject {
	text := widget.TextSegment{Style: widget.RichTextStyle{}, Text: decryptOptionExplanation}
	return container.NewVBox(
		container.NewCenter(widget.NewButton("Decrypt", func() { drawDecryptForm(w) })),
		widget.NewRichText(&text),
	)
}

func handleFileDrop(w fyne.Window, uris []fyne.URI) {
	// TODO: In contrast to the file picker, this supports multiple files!
	//       Create a zip archive when appropriate.

	if len(uris) == 1 {
		fileInfo, err := os.Stat(uris[0].Path())
		if err != nil {
			showMenuError(fmt.Errorf("Could not read file: %w", err), w)
			return
		} else if fileInfo.IsDir() {
			showMenuError(fmt.Errorf("spind cannot encrypt whole directories, only single files."), w)
			return
		}
		if droppedIn, err := storage.OpenFileFromURI(uris[0]); err == nil {
			in = droppedIn
			if strings.HasSuffix(strings.ToLower(in.URI().Path()), ".age") {
				drawDecryptForm(w)
			} else {
				drawEncryptForm(w)
			}
		} else {
			showMenuError(fmt.Errorf("Could not read file: %w", err), w)
		}
	} else {
		showMenuError(fmt.Errorf("Drop a single file into spind to en- or decrypt."), w)
	}
}
