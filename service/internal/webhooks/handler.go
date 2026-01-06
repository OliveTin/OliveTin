package webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/OliveTin/OliveTin/internal/auth"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	log "github.com/sirupsen/logrus"
)

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

	maxSize := int64(1024 * 1024)
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxSize))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warnf("Failed to read webhook payload")
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}

	var bodyData interface{}
	if err := json.Unmarshal(payload, &bodyData); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Debugf("Webhook payload is not valid JSON")
	}

	matchingActions := h.findMatchingActions(r, payload, bodyData)

	if len(matchingActions) == 0 {
		log.WithFields(log.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Debugf("No matching webhook actions found")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	processed := 0
	for _, actionConfig := range matchingActions {
		if h.processWebhook(actionConfig, r, payload) {
			processed++
		}
	}

	log.WithFields(log.Fields{
		"matched":   len(matchingActions),
		"processed": processed,
	}).Infof("Webhook processed")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *WebhookHandler) findMatchingActions(r *http.Request, payload []byte, bodyData interface{}) []ActionWebhookConfig {
	var matches []ActionWebhookConfig

	for _, action := range h.cfg.Actions {
		for _, webhookConfig := range action.ExecOnWebhook {
			webhookConfigCopy := webhookConfig

			if webhookConfigCopy.Template != "" {
				ApplyGitHubTemplate(&webhookConfigCopy, webhookConfigCopy.Template)
			}

			matcher := NewWebhookMatcher(webhookConfigCopy, r, payload, bodyData)

			if matcher.Matches() {
				matches = append(matches, ActionWebhookConfig{
					Action: action,
					Config: webhookConfigCopy,
				})
			}
		}
	}

	return matches
}

func (h *WebhookHandler) processWebhook(actionConfig ActionWebhookConfig, r *http.Request, payload []byte) bool {
	verifier := NewAuthVerifier(actionConfig.Config)
	if !verifier.Verify(r, payload) {
		log.WithFields(log.Fields{
			"actionTitle": actionConfig.Action.Title,
			"authType":    actionConfig.Config.AuthType,
		}).Warnf("Webhook authentication failed")
		return false
	}

	var bodyData interface{}
	json.Unmarshal(payload, &bodyData)

	matcher := NewWebhookMatcher(actionConfig.Config, r, payload, bodyData)

	args, err := matcher.ExtractArguments()
	if err != nil {
		log.WithFields(log.Fields{
			"actionTitle": actionConfig.Action.Title,
			"error":       err,
		}).Warnf("Failed to extract webhook arguments")
		return false
	}

	h.executeAction(actionConfig.Action, args)
	return true
}

func (h *WebhookHandler) executeAction(action *config.Action, args map[string]string) {
	binding := h.executor.FindBindingWithNoEntity(action)
	if binding == nil {
		log.WithFields(log.Fields{
			"actionTitle": action.Title,
		}).Warnf("Action binding not found, skipping execution")
		return
	}

	req := &executor.ExecutionRequest{
		Binding:           binding,
		Cfg:               h.cfg,
		Tags:              []string{"webhook"},
		Arguments:         args,
		AuthenticatedUser: auth.UserFromSystem(h.cfg, "webhook"),
	}

	h.executor.ExecRequest(req)
}
