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

func drawEncryptForm(w fyne.Window) {
	rawButton := widget.NewButtonWithIcon("Back to menu", theme.NavigateBackIcon(), func() { drawMenu(w) })
	menuButton := container.NewHBox(container.NewPadded(rawButton))
	content := container.New(layout.NewBorderLayout(menuButton, nil, nil, nil), menuButton, encryptForm(w))
	w.SetContent(content)
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
}

func encryptForm(w fyne.Window) fyne.CanvasObject {
	infileLabel := widget.NewLabel("File to encrypt")
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

	pw2Label := widget.NewLabel("Repeat password")
	pw2 := widget.NewPasswordEntry()
	pw2.Validator = func(s string) error {
		if s != pw.Text {
			return fmt.Errorf("password missmatch")
		}
		return nil
	}

	// FIXME: There's gotta be a better way than space-padding...
	pad := strings.Repeat(" ", 32)
	submit := widget.NewButton(pad+"Encrypt"+pad, func() { encrypt(pw, pw2, infile, w) })
	form := container.New(layout.NewFormLayout(), infileLabel, infile, pwLabel, pw, pw2Label, pw2)
	return container.NewPadded(container.NewCenter(container.NewVBox(form, submit)))
}

func encrypt(pw, pw2 *widget.Entry, infile *widget.Button, w fyne.Window) {
	if in == nil {
		showSubmenuError(fmt.Errorf("No input file selected."), w)
		return
	} else if pw.Text == "" {
		showSubmenuError(fmt.Errorf("Password is empty."), w)
		return
	} else if pw.Text != pw2.Text {
		showSubmenuError(fmt.Errorf("Passwords do not match."), w)
		return
	}
	nextDialogOpen := false
	fd := dialog.NewFileSave(func(out fyne.URIWriteCloser, err error) {
		nextDialogOpen = true // User didn't abort.
		if err != nil {
			showSubmenuError(fmt.Errorf("Could not write new file: %w.", err), w)
			return
		}
		if out != nil {
			defer out.Close()
			defer func() {
				in.Close()
				in = nil
				infile.SetText("Select input file")
			}()
			if err := encryptToFile(out, pw.Text, w); err != nil {
				showSubmenuError(fmt.Errorf("Could not encrypt file: %w.", err), w)
				return
			}
		}
	}, w)
	if in != nil {
		fd.SetFileName(in.URI().Name() + ".age")
		if parent, err := storage.Parent(in.URI()); err == nil {
			if l, err := storage.ListerForURI(parent); err == nil {
				fd.SetLocation(l)
			}
		}
	}
	fd.Resize(dialogSize)
	fd.SetOnClosed(func() {
		if !nextDialogOpen {
			w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleSubmenuKey(e, w) })
		}
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, fd) })
	fd.Show()
}

func encryptToFile(out fyne.URIWriteCloser, pw string, w fyne.Window) error {
	recipient, err := age.NewScryptRecipient(pw)
	if err != nil {
		return err
	}
	writer, err := age.Encrypt(out, recipient)
	if err != nil {
		return err
	}
	defer writer.Close()
	txt := fmt.Sprintf("Encrypting %s...", in.URI().Path())
	if totalSize := sizeOf(in); totalSize > 0 {
		progressDialog := dialog.NewProgress("Encrypting", txt, w)
		w.Canvas().SetOnTypedKey(nil) // Keys will be enabled again in success or error dialog.
		progressDialog.Show()
		_, err = io.Copy(&progressWriter{writer, totalSize, 0, progressDialog, time.Now()}, in)
		progressDialog.Hide()
		if err != nil {
			return err
		}
	} else {
		progressDialog := dialog.NewProgressInfinite("Encrypting", txt, w)
		w.Canvas().SetOnTypedKey(nil) // Keys will be enabled again in success or error dialog.
		progressDialog.Show()
		_, err = io.Copy(writer, in)
		progressDialog.Hide()
		if err != nil {
			return err
		}
	}
	showEncryptionSuccess(in.URI(), out.URI(), w)
	return nil
}

func showEncryptionSuccess(in, out fyne.URI, w fyne.Window) {
	inLabel := widget.NewLabel("Input was:")
	inPath := widget.NewLabel(in.Path())
	outLabel := widget.NewLabel("New file:")
	outPath := widget.NewLabel(out.Path())
	body := container.New(layout.NewFormLayout(), inLabel, inPath, outLabel, outPath) // FIXME: Reduce space between rows.

	d := dialog.NewCustom("Encryption Successful", "OK", body, w)
	d.SetOnClosed(func() { drawMenu(w) })
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) { handleDialogKey(e, w, d) })
	d.Show()
}
