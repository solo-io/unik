package util

import (
	"gopkg.in/cheggaaa/pb.v1"
	"io"
)

///http://stackoverflow.com/questions/22421375/how-to-print-the-bytes-while-the-file-is-being-downloaded-golang
// WriteCounter counts the number of bytes written to it.
type writeCounter struct {
	current int64 // Total # of bytes transferred
	total   int64 // Expected length
	bar     *pb.ProgressBar
}

func newWriteCounter(total int64) *writeCounter {
	return &writeCounter{
		total: total,
		bar:   pb.StartNew(int(total)),
	}
}

// Write implements the io.Writer interface.
//
// Always completes and never returns an error.
func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += int64(n)
	wc.bar.Set(int(wc.current))
	if wc.current >= wc.total-1 {
		wc.bar.FinishPrint("download complete")
	}
	return n, nil
}

func ReaderWithProgress(r io.Reader, total int64) io.Reader {
	return io.TeeReader(r, newWriteCounter(total))
}
