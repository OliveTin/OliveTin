package httpservers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	localUserSessions = make(map[string]string) // sid -> username, used for local user sessions
)

func parseLocalUserCookie(req *http.Request) (string, string) {
	cookie, err := req.Cookie("olivetin_local_user_sid")
	if err != nil {
		return "", ""
	}

	cookieValue := cookie.Value

	username, ok := localUserSessions[cookieValue]

	if !ok {
		log.Warnf("Could not find local user session: %v", cookieValue)
		return "", ""
	}

	return username, ""
}

func forwardResponseHandlerLoginLocalUser(ctx context.Context, w http.ResponseWriter, msg protoreflect.ProtoMessage) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)

	if !ok {
		log.Warn("Could not get ServerMetadata from context")
		return nil
	}

	setUser := acl.SetUserFromMetadata(md.HeaderMD)

	sid := uuid.NewString()
	localUserSessions[sid] = setUser

	if setUser != "" {
		http.SetCookie(
			w,
			&http.Cookie{
				Name:  "olivetin_local_user_sid",
				Value: sid,
			},
		)
	}

	return nil
}
