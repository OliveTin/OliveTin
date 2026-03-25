package api

import (
	"net/http"

	executor "github.com/OliveTin/OliveTin/internal/executor"
)

// GetActionArgumentUploadHandler serves POST multipart uploads for file_upload action arguments.
func GetActionArgumentUploadHandler(ex *executor.Executor) http.HandlerFunc {
	api := &oliveTinAPI{
		executor:         ex,
		cfg:              ex.Cfg,
		streamingClients: make(map[*streamingClient]struct{}),
	}
	return api.handleActionArgumentUpload
}

func (api *oliveTinAPI) handleActionArgumentUpload(w http.ResponseWriter, r *http.Request) {
	if !api.uploadPrelude(w, r) {
		return
	}
	if !api.parseUploadForm(w, r) {
		return
	}
	defer api.uploadCleanupForm(r)

	token, ok := api.tryProcessUpload(w, r)
	if !ok {
		return
	}
	api.writeUploadTokenResponse(w, token)
}
