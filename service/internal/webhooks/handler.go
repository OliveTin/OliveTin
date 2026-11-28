package webhooks

import (
	"io"
	"net/http"

	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// CorrelationIDHeader is the response header carrying the ID that appears in
// OliveTin's log lines for the same webhook request.
const CorrelationIDHeader = "X-OliveTin-Log-Correlation-ID"

type ActionWebhookConfig struct {
	Action *config.Action
	Config config.WebhookConfig
}

type WebhookHandler struct {
	cfg      *config.Config
	executor *executor.Executor
}

func NewWebhookHandler(cfg *config.Config, ex *executor.Executor) *WebhookHandler {
	return &WebhookHandler{
		cfg:      cfg,
		executor: ex,
	}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	correlationID := uuid.NewString()
	w.Header().Set(CorrelationIDHeader, correlationID)
	reqLog := log.WithFields(webhookRequestFields(r, correlationID))

	payload, err := h.readPayload(r, reqLog)
	if err != nil {
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}

	matchingActions := h.findMatchingActions(r, payload, reqLog)
	if len(matchingActions) == 0 {
		reqLog.Infof("Webhook: no matching actions")
		h.writeOKResponse(w, "no matching webhook actions")
		return
	}

	processed := h.processMatchingActions(matchingActions, r, payload, reqLog)
	reqLog.WithFields(log.Fields{
		"matched":   len(matchingActions),
		"processed": processed,
	}).Infof("Webhook: matched")

	h.writeOKResponse(w, "webhook actions")
}

func webhookRequestFields(r *http.Request, correlationID string) log.Fields {
	return log.Fields{
		"correlationId": correlationID,
		"method":        r.Method,
		"path":          r.URL.Path,
		"query":         r.URL.RawQuery,
	}
}

func (h *WebhookHandler) readPayload(r *http.Request, reqLog *log.Entry) ([]byte, error) {
	maxSize := int64(1024 * 1024)
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxSize))
	if err != nil {
		reqLog.WithFields(log.Fields{
			"error": err,
		}).Warnf("Failed to read webhook payload")
		return nil, err
	}

	return payload, nil
}

func (h *WebhookHandler) writeOKResponse(w http.ResponseWriter, context string) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.WithError(err).Warnf("Failed to write response for %s", context)
	}
}

func (h *WebhookHandler) processMatchingActions(matchingActions []ActionWebhookConfig, r *http.Request, payload []byte, reqLog *log.Entry) int {
	processed := 0
	for _, actionConfig := range matchingActions {
		if h.processWebhook(actionConfig, r, payload, reqLog) {
			processed++
		}
	}
	return processed
}

func (h *WebhookHandler) findMatchingActions(r *http.Request, payload []byte, reqLog *log.Entry) []ActionWebhookConfig {
	var matches []ActionWebhookConfig

	for _, action := range h.cfg.Actions {
		matches = append(matches, h.findMatchingWebhooksForAction(action, r, payload, reqLog)...)
	}

	return matches
}

func (h *WebhookHandler) findMatchingWebhooksForAction(action *config.Action, r *http.Request, payload []byte, reqLog *log.Entry) []ActionWebhookConfig {
	var matches []ActionWebhookConfig

	for _, webhookConfig := range action.ExecOnWebhook {
		webhookConfigCopy := webhookConfig

		if webhookConfigCopy.Template != "" {
			ApplyGitHubTemplate(&webhookConfigCopy, webhookConfigCopy.Template)
		}

		matcher := NewWebhookMatcher(webhookConfigCopy, r, payload).WithLogger(reqLog)
		if matcher.Matches() {
			matches = append(matches, ActionWebhookConfig{
				Action: action,
				Config: webhookConfigCopy,
			})
		}
	}

	return matches
}

func (h *WebhookHandler) processWebhook(actionConfig ActionWebhookConfig, r *http.Request, payload []byte, reqLog *log.Entry) bool {
	verifier := NewAuthVerifier(actionConfig.Config)
	if !verifier.Verify(r, payload) {
		reqLog.WithFields(log.Fields{
			"actionTitle": actionConfig.Action.Title,
			"authType":    actionConfig.Config.AuthType,
		}).Warnf("Webhook authentication failed")
		return false
	}

	matcher := NewWebhookMatcher(actionConfig.Config, r, payload).WithLogger(reqLog)

	args, err := matcher.ExtractArguments()
	if err != nil {
		reqLog.WithFields(log.Fields{
			"actionTitle": actionConfig.Action.Title,
			"error":       err,
		}).Warnf("Failed to extract webhook arguments")
		return false
	}

	justification, err := matcher.ExtractJustification()
	if err != nil {
		reqLog.WithFields(log.Fields{
			"actionTitle": actionConfig.Action.Title,
			"error":       err,
		}).Warnf("Failed to extract webhook justification")
		return false
	}

	h.executeAction(actionConfig.Action, args, justification, reqLog)
	return true
}

func (h *WebhookHandler) executeAction(action *config.Action, args map[string]string, justification string, reqLog *log.Entry) {
	binding := h.executor.FindBindingWithNoEntity(action)
	if binding == nil {
		reqLog.WithFields(log.Fields{
			"actionTitle": action.Title,
		}).Warnf("Action binding not found, skipping execution")
		return
	}

	definedArgs := filterToDefinedArguments(args, action)
	req := &executor.ExecutionRequest{
		Binding:           binding,
		Cfg:               h.cfg,
		Tags:              []string{"webhook"},
		Arguments:         definedArgs,
		Justification:     justification,
		AuthenticatedUser: auth.UserFromSystem(h.cfg, "webhook"),
	}

	h.executor.ExecRequest(req)
}

func filterToDefinedArguments(args map[string]string, action *config.Action) map[string]string {
	definedNames := make(map[string]struct{})
	for _, arg := range action.Arguments {
		definedNames[arg.Name] = struct{}{}
	}
	filtered := make(map[string]string)
	for k, v := range args {
		if _, ok := definedNames[k]; ok {
			filtered[k] = v
		}
	}
	return filtered
}
