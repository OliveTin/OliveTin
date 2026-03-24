package api

import (
	"encoding/json"
	"mime/multipart"
	"net/http"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	auth "github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	executor "github.com/OliveTin/OliveTin/internal/executor"
	log "github.com/sirupsen/logrus"
)

func (api *oliveTinAPI) uploadPrelude(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	if api.executor.UploadRegistry == nil {
		http.Error(w, "file uploads are not available", http.StatusServiceUnavailable)
		return false
	}
	return true
}

func uploadMaxBodyBytes(cfg *config.Config) int64 {
	maxBody := int64(cfg.FileUploads.MaxBytes)
	if maxBody <= 0 {
		maxBody = int64(config.DefaultFileUploadMaxBytes)
	}
	return maxBody
}

func (api *oliveTinAPI) parseUploadForm(w http.ResponseWriter, r *http.Request) bool {
	maxBody := uploadMaxBodyBytes(api.cfg)
	r.Body = http.MaxBytesReader(w, r.Body, maxBody+(1<<20))
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return false
	}
	return true
}

func uploadFormIDs(r *http.Request) (string, string, bool) {
	bindingID := r.FormValue("binding_id")
	argName := r.FormValue("argument_name")
	if bindingID == "" || argName == "" {
		return "", "", false
	}
	return bindingID, argName, true
}

func (api *oliveTinAPI) uploadCleanupForm(r *http.Request) {
	if r.MultipartForm != nil {
		_ = r.MultipartForm.RemoveAll()
	}
}

func (api *oliveTinAPI) bindingForUpload(w http.ResponseWriter, bindingID string) *executor.ActionBinding {
	pair := api.executor.FindBindingByID(bindingID)
	if pair == nil || pair.Action == nil {
		http.Error(w, "action not found", http.StatusNotFound)
		return nil
	}
	return pair
}

func (api *oliveTinAPI) authorizeUploadRequest(w http.ResponseWriter, r *http.Request, bindingID, argName string) (*executor.ActionBinding, *config.ActionArgument) {
	user := auth.UserFromHTTPRequest(r, api.cfg)
	pair := api.bindingForUpload(w, bindingID)
	if pair == nil {
		return nil, nil
	}
	if !uploadExecAllowed(api, user, pair) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, nil
	}
	return uploadFileArgOrError(api, w, pair, argName)
}

func uploadExecAllowed(api *oliveTinAPI, user *authpublic.AuthenticatedUser, pair *executor.ActionBinding) bool {
	return acl.IsAllowedExec(api.cfg, user, pair.Action)
}

func uploadFileArgOrError(api *oliveTinAPI, w http.ResponseWriter, pair *executor.ActionBinding, argName string) (*executor.ActionBinding, *config.ActionArgument) {
	arg := api.findArgumentByName(pair.Action, argName)
	if arg == nil || arg.Type != "file_upload" {
		http.Error(w, "invalid file argument", http.StatusBadRequest)
		return nil, nil
	}
	return pair, arg
}

func (api *oliveTinAPI) openUploadedFormFile(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, bool) {
	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return nil, nil, false
	}
	return file, hdr, true
}

func (api *oliveTinAPI) tryProcessUpload(w http.ResponseWriter, r *http.Request) (string, bool) {
	bindingID, argName, ok := uploadFormIDs(r)
	if !ok {
		http.Error(w, "binding_id and argument_name are required", http.StatusBadRequest)
		return "", false
	}
	_, arg := api.authorizeUploadRequest(w, r, bindingID, argName)
	if arg == nil {
		return "", false
	}
	return api.stageUploadFromForm(w, r, bindingID, arg)
}

func (api *oliveTinAPI) stageUploadFromForm(w http.ResponseWriter, r *http.Request, bindingID string, arg *config.ActionArgument) (string, bool) {
	file, hdr, ok := api.openUploadedFormFile(w, r)
	if !ok {
		return "", false
	}
	defer file.Close()
	token, err := api.executor.UploadRegistry.StageFromMultipart(file, hdr.Filename, bindingID, arg)
	if err != nil {
		log.WithError(err).Warn("upload rejected")
		http.Error(w, "upload rejected", http.StatusBadRequest)
		return "", false
	}
	return token, true
}

func (api *oliveTinAPI) writeUploadTokenResponse(w http.ResponseWriter, token string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"uploadToken": token}); err != nil {
		log.WithError(err).Warn("upload response encode failed")
	}
}
