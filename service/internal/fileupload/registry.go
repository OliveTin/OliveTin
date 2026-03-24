package fileupload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	config "github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

// StagedFile is a validated upload ready for template expansion and command execution.
type StagedFile struct {
	Path         string
	OriginalName string
	MimeType     string
	Size         int64
}

// Registry stores single-use upload tokens mapped to temp files on disk.
type Registry struct {
	mu      sync.Mutex
	pending map[string]*pendingEntry
	cfg     *config.Config
	baseDir string
}

type pendingEntry struct {
	path         string
	bindingID    string
	argName      string
	originalName string
	mimeType     string
	size         int64
	expires      time.Time
}

// NewRegistry creates an upload registry and ensures the staging directory exists.
func NewRegistry(cfg *config.Config) (*Registry, error) {
	if cfg == nil {
		return nil, fmt.Errorf("fileupload: config is nil")
	}
	abs, err := resolveUploadBaseDir(cfg)
	if err != nil {
		return nil, fmt.Errorf("fileupload: temp directory: %w", err)
	}
	return &Registry{
		cfg:     cfg,
		pending: make(map[string]*pendingEntry),
		baseDir: abs,
	}, nil
}

// StartPeriodicPrune runs a background loop that removes pending uploads past their TTL.
// Without this, staged files are only deleted when some other registry operation runs prune.
func (r *Registry) StartPeriodicPrune() {
	go r.periodicPruneLoop()
}

func (r *Registry) periodicPruneLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		r.pruneExpired()
	}
}

func (r *Registry) pruneExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pruneLocked()
}

func resolveUploadBaseDir(cfg *config.Config) (string, error) {
	base := cfg.FileUploads.TempDirectory
	if base == "" {
		base = filepath.Join(os.TempDir(), "olivetin-uploads")
	}
	abs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return "", err
	}
	return abs, nil
}

func (r *Registry) tokenTTL() time.Duration {
	sec := r.cfg.FileUploads.TokenTTLSeconds
	if sec <= 0 {
		sec = config.DefaultFileUploadTokenTTLSeconds
	}
	return time.Duration(sec) * time.Second
}

// StageFromMultipart saves the body to a private temp file, validates MIME type, and returns an opaque token.
func (r *Registry) StageFromMultipart(
	file io.Reader,
	filenameHint string,
	bindingID string,
	arg *config.ActionArgument,
) (string, error) {
	if err := validateStageArgument(arg); err != nil {
		return "", err
	}
	maxBytes := arg.EffectiveFileUploadMaxBytes(r.cfg)
	allowed := arg.EffectiveFileUploadAllowedMimeTypes(r.cfg)
	if len(allowed) == 0 {
		return "", fmt.Errorf("no allowedMimeTypes configured for this argument (configure argument or fileUploads.defaultAllowedMimeTypes)")
	}
	return r.stageMultipartBody(file, filenameHint, bindingID, arg, maxBytes, allowed)
}

func (r *Registry) stageMultipartBody(
	file io.Reader,
	filenameHint, bindingID string,
	arg *config.ActionArgument,
	maxBytes int64,
	allowed []string,
) (string, error) {
	tmpPath, n, err := r.copyLimitedToTemp(file, maxBytes)
	if err != nil {
		return "", err
	}
	if arg.RejectNull && n == 0 {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("empty file not allowed")
	}
	return r.finishStagedFile(tmpPath, filenameHint, bindingID, arg, n, allowed)
}

func (r *Registry) finishStagedFile(
	tmpPath, filenameHint, bindingID string,
	arg *config.ActionArgument,
	n int64,
	allowed []string,
) (string, error) {
	detected, err := detectMimeFromPath(tmpPath)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if !mimeAllowed(detected, allowed) {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("MIME type %q is not allowed", detected)
	}
	token, err := r.storeStagedUpload(tmpPath, bindingID, filenameHint, arg, detected, n)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	log.WithFields(log.Fields{
		"bindingId": bindingID,
		"arg":       arg.Name,
		"mime":      detected,
		"bytes":     n,
	}).Debug("staged file upload")
	return token, nil
}

func (r *Registry) storeStagedUpload(tmpPath, bindingID, filenameHint string, arg *config.ActionArgument, detected string, n int64) (string, error) {
	token, err := newUploadToken()
	if err != nil {
		return "", err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pruneLocked()
	r.pending[token] = &pendingEntry{
		path:         tmpPath,
		bindingID:    bindingID,
		argName:      arg.Name,
		originalName: sanitizeFilename(filenameHint),
		mimeType:     detected,
		size:         n,
		expires:      time.Now().Add(r.tokenTTL()),
	}
	return token, nil
}

func validateStageArgument(arg *config.ActionArgument) error {
	if arg == nil || arg.Type != "file_upload" {
		return fmt.Errorf("invalid file upload argument")
	}
	return nil
}

func (r *Registry) pruneLocked() {
	now := time.Now()
	for k, v := range r.pending {
		if now.After(v.expires) {
			_ = os.Remove(v.path)
			delete(r.pending, k)
		}
	}
}

// ValidatePeekToken checks that a token exists and matches the binding and argument (does not consume).
func (r *Registry) ValidatePeekToken(token, bindingID, argName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pruneLocked()
	return r.peekLocked(token, bindingID, argName)
}

func (r *Registry) peekLocked(token, bindingID, argName string) error {
	ent, ok := r.pending[token]
	if !ok {
		return fmt.Errorf("unknown or expired upload token")
	}
	if r.pendingEntryExpired(ent, token) {
		return fmt.Errorf("unknown or expired upload token")
	}
	return r.pendingEntryMatches(ent, bindingID, argName)
}

func (r *Registry) pendingEntryExpired(ent *pendingEntry, token string) bool {
	if !time.Now().After(ent.expires) {
		return false
	}
	_ = os.Remove(ent.path)
	delete(r.pending, token)
	return true
}

func (r *Registry) pendingEntryMatches(ent *pendingEntry, bindingID, argName string) error {
	if ent.bindingID != bindingID || ent.argName != argName {
		return fmt.Errorf("upload token does not match this action argument")
	}
	if !r.fileWithinBase(ent.path) {
		return fmt.Errorf("invalid staged file path")
	}
	return nil
}

// ConsumeToken removes the token and returns the staged file (single use).
func (r *Registry) ConsumeToken(token, bindingID, argName string) (*StagedFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pruneLocked()
	if err := r.peekLocked(token, bindingID, argName); err != nil {
		return nil, err
	}
	ent := r.pending[token]
	delete(r.pending, token)
	return &StagedFile{
		Path:         ent.path,
		OriginalName: ent.originalName,
		MimeType:     ent.mimeType,
		Size:         ent.size,
	}, nil
}

// DeleteTempFile removes a temp file if it resides under the registry base directory.
func (r *Registry) DeleteTempFile(path string) {
	if !r.fileWithinBase(path) {
		log.Warnf("refusing to delete path outside upload base: %s", path)
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Warnf("remove upload temp file: %v", err)
	}
}

func (r *Registry) fileWithinBase(path string) bool {
	absFile, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(r.baseDir, absFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

func newUploadToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	if base == "." || base == string(filepath.Separator) {
		return "upload"
	}
	if len(base) > 255 {
		base = base[:255]
	}
	return base
}
