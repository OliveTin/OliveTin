package executor

import (
	"fmt"
	"regexp"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/fileupload"
	"github.com/OliveTin/OliveTin/internal/tpl"
)

var fileUploadTokenPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func validateFileUploadArg(value string, arg *config.ActionArgument, reg *fileupload.Registry, bindingID string) error {
	if value == "" {
		return typecheckNull(arg)
	}
	if !fileUploadTokenPattern.MatchString(value) {
		return fmt.Errorf("invalid upload token")
	}
	if reg == nil {
		return errUploadsUnavailable()
	}
	return reg.ValidatePeekToken(value, bindingID, arg.Name)
}

func finalizeFileUploadArguments(req *ExecutionRequest) error {
	if !hasActionForFileFinalize(req) {
		return nil
	}
	if req.FileArgData == nil {
		req.FileArgData = make(map[string]*tpl.FileUpload)
	}
	return finalizeEachFileUploadArg(req)
}

func finalizeEachFileUploadArg(req *ExecutionRequest) error {
	for i := range req.Binding.Action.Arguments {
		arg := &req.Binding.Action.Arguments[i]
		if arg.Type != "file_upload" {
			continue
		}
		if err := finalizeOneFileUpload(req, arg); err != nil {
			return err
		}
	}
	return nil
}

func hasActionForFileFinalize(req *ExecutionRequest) bool {
	return req != nil && req.Binding != nil && req.Binding.Action != nil
}

func finalizeOneFileUpload(req *ExecutionRequest, arg *config.ActionArgument) error {
	raw := req.Arguments[arg.Name]
	if raw == "" {
		return finalizeEmptyFileArg(req, arg)
	}
	reg := req.executor.UploadRegistry
	if reg == nil {
		return errUploadsUnavailable()
	}
	staged, err := reg.ConsumeToken(raw, req.Binding.ID, arg.Name)
	if err != nil {
		return err
	}
	applyConsumedStagedFile(req, arg, staged)
	return nil
}

func finalizeEmptyFileArg(req *ExecutionRequest, arg *config.ActionArgument) error {
	if arg.RejectNull {
		return errRejectNullFile(arg.Name)
	}
	req.FileArgData[arg.Name] = nil
	return nil
}

func applyConsumedStagedFile(req *ExecutionRequest, arg *config.ActionArgument, staged *fileupload.StagedFile) {
	req.UploadTempPaths = append(req.UploadTempPaths, staged.Path)
	req.Arguments[arg.Name] = staged.Path
	req.FileArgData[arg.Name] = &tpl.FileUpload{
		TmpName:  staged.Path,
		Name:     fileupload.SanitizeUploadFilename(staged.OriginalName),
		MimeType: staged.MimeType,
		Size:     staged.Size,
	}
}

func buildTemplateArgumentMap(req *ExecutionRequest) map[string]any {
	out := make(map[string]any)
	for k, v := range req.Arguments {
		if fu, ok := req.FileArgData[k]; ok {
			out[k] = fu
			continue
		}
		out[k] = v
	}
	return out
}

func triggerArgumentsWithoutUploads(req *ExecutionRequest) map[string]string {
	if !hasBindingAndAction(req) {
		return nil
	}
	out := make(map[string]string, len(req.Arguments))
	for k, v := range req.Arguments {
		out[k] = v
	}
	clearFileUploadArgs(out, req.Binding.Action.Arguments)
	return out
}

func clearFileUploadArgs(out map[string]string, args []config.ActionArgument) {
	for i := range args {
		if args[i].Type == "file_upload" {
			out[args[i].Name] = ""
		}
	}
}

func errRejectNullFile(name string) error {
	return fmt.Errorf("argument %s requires a file", name)
}

func errUploadsUnavailable() error {
	return fmt.Errorf("file uploads are not available on this server")
}
