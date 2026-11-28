package httpservers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// A replacement web UI file can keep the same, or an older, modification time
// as the file it replaces (for example when it is copied with its times
// preserved), so Last-Modified cannot tell a browser that it changed. Files are
// validated by a hash of their content instead, and no modification time is
// sent or compared, so a browser still holding a Last-Modified value from an
// older release always gets the current file.

var errServeIsDirectory = errors.New("path is a directory")

func serveFileWithContentETag(w http.ResponseWriter, r *http.Request, root http.FileSystem, name string, fallback http.Handler) {
	file, info, err := openServableFile(root, name)
	if err != nil {
		fallback.ServeHTTP(w, r)
		return
	}

	defer closeServedFile(file, name)

	etag, err := contentETag(file)
	if err != nil {
		log.Warnf("Could not hash web UI file %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("ETag", etag)
	http.ServeContent(w, r, info.Name(), time.Time{}, file)
}

func openServableFile(root http.FileSystem, name string) (http.File, fs.FileInfo, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err == nil && info.IsDir() {
		err = errServeIsDirectory
	}

	if err != nil {
		closeServedFile(file, name)
		return nil, nil, err
	}

	return file, info, nil
}

func closeServedFile(file http.File, name string) {
	if err := file.Close(); err != nil {
		log.Warnf("Could not close web UI file %s: %v", name, err)
	}
}

func contentETag(file http.File) (string, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return `"` + hex.EncodeToString(hash.Sum(nil)[:16]) + `"`, nil
}
