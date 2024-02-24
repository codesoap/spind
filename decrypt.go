package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"filippo.io/age"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func drawDecryptForm(w fyne.Window) {
	rawButton := widget.NewButtonWithIcon("Back to menu", theme.NavigateBackIcon(), func() { drawMenu(w) })
	menuButton := container.NewHBox(container.NewPadded(rawButton))
	content := container.New(layout.NewBorderLayout(menuButton, nil, nil, nil), menuButton, decryptForm(w))
	w.SetContent(content)
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
}

func decryptForm(w fyne.Window) fyne.CanvasObject {
	infileLabel := widget.NewLabel("File to decrypt")
	var infile *widget.Button
	if in != nil {
		// File provided via drag and drop.
		infile = widget.NewButton(in.URI().Name(), nil)
	} else {
		infile = widget.NewButton("Select input file", nil)
	}
	infile.OnTapped = func() { fillInfile(infile, w) }

	pwLabel := widget.NewLabel("Password")
	pw := widget.NewPasswordEntry()
	pw.Validator = stringEmpty

	// FIXME: There's gotta be a better way than space-padding...
	pad := strings.Repeat(" ", 32)
	submit := widget.NewButton(pad+"Decrypt"+pad, func() { decrypt(pw, infile, w) })
	form := container.New(layout.NewFormLayout(), infileLabel, infile, pwLabel, pw)
	return container.NewPadded(container.NewCenter(container.NewVBox(form, submit)))
}

func decrypt(pw *widget.Entry, infile *widget.Button, w fyne.Window) {
	if in == nil {
		showError(fmt.Errorf("No input file selected."), w)
		return
	} else if pw.Text == "" {
		showError(fmt.Errorf("Password is empty."), w)
		return
	}
	nextDialogOpen := false
	fd := dialog.NewFileSave(func(out fyne.URIWriteCloser, err error) {
		nextDialogOpen = true // User didn't abort.
		if err != nil {
			showError(fmt.Errorf("Could not write new file: %w.", err), w)
			return
		}
		if out != nil {
			defer out.Close()
			defer func() {
				in.Close()
				in = nil
				infile.SetText("Select input file")
			}()
			if err := decryptToFile(out, pw.Text, w); err != nil {
				showError(fmt.Errorf("Could not decrypt file: %w.", err), w)
				return
			}
		}
	}, w)
	if in != nil {
		filename := in.URI().Name()
		if strings.HasSuffix(filename, ".age") {
			fd.SetFileName(filename[:len(filename)-4])
		}
		if parent, err := storage.Parent(in.URI()); err == nil {
			if l, err := storage.ListerForURI(parent); err == nil {
				fd.SetLocation(l)
			}
		}
	}
	fd.Resize(dialogSize)
	fd.Show() // FIXME: Move down once #4651 is fixed.
	fd.SetOnClosed(func() {
		if !nextDialogOpen {
			w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
		}
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, fd) })
}

func decryptToFile(out fyne.URIWriteCloser, pw string, w fyne.Window) error {
	identity, err := age.NewScryptIdentity(pw)
	if err != nil {
		return err
	}
	reader, err := age.Decrypt(in, identity)
	if _, ok := err.(*age.NoIdentityMatchError); ok {
		return fmt.Errorf("wrong password")
	} else if err != nil {
		return err
	}
	txt := fmt.Sprintf("Decrypting %s...", in.URI().Path())
	if totalSize := sizeOf(in); totalSize > 0 {
		progressDialog := dialog.NewProgress("Decrypting", txt, w)
		w.Canvas().SetOnTypedKey(nil) // Keys will be enabled again in success or error dialog.
		progressDialog.Show()
		_, err = io.Copy(out, &progressReader{reader, totalSize, 0, progressDialog, time.Now()})
		progressDialog.Hide()
		if err != nil {
			return err
		}
	} else {
		progressDialog := dialog.NewProgressInfinite("Decrypting", txt, w)
		w.Canvas().SetOnTypedKey(nil) // Keys will be enabled again in success or error dialog.
		progressDialog.Show()
		_, err = io.Copy(out, reader)
		progressDialog.Hide()
		if err != nil {
			return err
		}
	}
	showDecryptionSuccess(in.URI(), out.URI(), w)
	return nil
}

func showDecryptionSuccess(in, out fyne.URI, w fyne.Window) {
	inLabel := widget.NewLabel("Input was:")
	inPath := widget.NewLabel(in.Path())
	outLabel := widget.NewLabel("New file:")
	outPath := widget.NewLabel(out.Path())
	body := container.New(layout.NewFormLayout(), inLabel, inPath, outLabel, outPath) // FIXME: Reduce space between rows.

	d := dialog.NewCustom("Decryption Successful", "OK", body, w)
	d.SetOnClosed(func() { drawMenu(w) })
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, d) })
	d.Show()
}
