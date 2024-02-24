package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func handleSubmenuKey(e *fyne.KeyEvent, w fyne.Window) {
	switch e.Name {
	case fyne.KeyEscape:
		drawMenu(w)
	case fyne.KeyQ:
		w.Close()
	}
}

func handleDialogKey(e *fyne.KeyEvent, w fyne.Window, d dialog.Dialog) {
	switch e.Name {
	case fyne.KeyEscape:
		d.Hide()
	case fyne.KeyQ:
		w.Close()
	}
}

func fillInfile(button *widget.Button, w fyne.Window) {
	fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			showSubmenuError(fmt.Errorf("Could not open file: %w.", err), w)
			return
		}
		if rc != nil {
			if in != nil {
				in.Close()
			}
			in = rc
			button.SetText(rc.URI().Name())
		} else {
			button.SetText("Select input file")
		}
	}, w)
	fd.Resize(dialogSize)
	fd.Show() // FIXME: Move down once #4651 is fixed.
	fd.SetOnClosed(func() {
		w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, fd) })
}

func showMenuError(err error, w fyne.Window) {
	errDialog := dialog.NewError(err, w)
	errDialog.Show() // FIXME: Move down once #4651 is fixed.
	errDialog.SetOnClosed(func() {
		w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleMenuKey(e, w) })
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, errDialog) })
}

func showSubmenuError(err error, w fyne.Window) {
	errDialog := dialog.NewError(err, w)
	errDialog.Show() // FIXME: Move down once #4651 is fixed.
	errDialog.SetOnClosed(func() {
		w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, errDialog) })
}

// sizeOf returns the size of the file of uri. If URI is not a file or
// the size cannot be determined for other reasons, 0 will be returned.
func sizeOf(uri fyne.URIReadCloser) int64 {
	if uri.URI().Scheme() == "file" {
		fileInfo, err := os.Stat(uri.URI().Path())
		if err == nil {
			return fileInfo.Size()
		}
	}
	return 0
}

func stringEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("string is empty")
	}
	return nil
}
