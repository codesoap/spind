package main

import (
	"io"
	"time"

	"fyne.io/fyne/v2/dialog"
)

type progressReader struct {
	io.Reader
	totalBytes     int64
	readBytes      int64
	progressDialog *dialog.ProgressDialog
	lastUpdate     time.Time
}

func (w *progressReader) Read(p []byte) (n int, err error) {
	n, err = w.Reader.Read(p)
	if err != nil {
		return n, err
	}
	w.readBytes += int64(n)
	if w.totalBytes != 0 && time.Now().Sub(w.lastUpdate) > 50*time.Millisecond {
		w.progressDialog.SetValue(float64(w.readBytes) / float64(w.totalBytes))
		w.progressDialog.Refresh()
		w.lastUpdate = time.Now()
	}
	return n, err
}
