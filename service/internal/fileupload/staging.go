package fileupload

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (r *Registry) copyLimitedToTemp(file io.Reader, maxBytes int64) (string, int64, error) {
	tmp, err := os.CreateTemp(r.baseDir, "olu-")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	path := tmp.Name()
	n, err := io.Copy(tmp, io.LimitReader(file, maxBytes+1))
	cerr := tmp.Close()
	return finalizeLimitedCopy(path, n, err, cerr, maxBytes)
}

func finalizeLimitedCopy(path string, n int64, err, cerr error, maxBytes int64) (string, int64, error) {
	if err != nil {
		_ = os.Remove(path)
		return "", 0, fmt.Errorf("read upload: %w", err)
	}
	if cerr != nil {
		_ = os.Remove(path)
		return "", 0, fmt.Errorf("close temp file: %w", cerr)
	}
	if n > maxBytes {
		_ = os.Remove(path)
		return "", n, fmt.Errorf("file exceeds maximum size of %d bytes", maxBytes)
	}
	return path, n, nil
}

func detectMimeFromPath(path string) (string, error) {
	header := make([]byte, 512)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	readN, readErr := f.Read(header)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", fmt.Errorf("read sniff buffer: %w", readErr)
	}
	return http.DetectContentType(header[:readN]), nil
}
