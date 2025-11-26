package authpublic

import (
	"net/http"

	"github.com/OliveTin/OliveTin/internal/config"
)

type AuthCheckingContext struct {
	Config  *config.Config
	Request *http.Request
}
