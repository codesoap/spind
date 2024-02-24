package main

import (
	"io"
	"time"

	"fyne.io/fyne/v2/dialog"
)

type progressWriter struct {
	io.Writer
	totalBytes     int64
	writtenBytes   int64
	progressDialog *dialog.ProgressDialog
	lastUpdate     time.Time
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		return n, err
	}
	w.writtenBytes += int64(n)
	if w.totalBytes != 0 && time.Now().Sub(w.lastUpdate) > 50*time.Millisecond {
		w.progressDialog.SetValue(float64(w.writtenBytes) / float64(w.totalBytes))
		w.progressDialog.Refresh()
		w.lastUpdate = time.Now()
	}
	return n, err
}
