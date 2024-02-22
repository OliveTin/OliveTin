package grpcapi

import (
	"crypto/sha256"
	"fmt"
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type ActionWithEntity struct {
	Action       *config.Action
	EntityPrefix string
}

var publicActionIdToActionMap map[string]ActionWithEntity

func init() {
	publicActionIdToActionMap = make(map[string]ActionWithEntity)
}

func actionsCfgToPb(cfgActions []*config.Action, user *acl.AuthenticatedUser) *pb.GetDashboardComponentsResponse {
	res := &pb.GetDashboardComponentsResponse{}

	for _, action := range cfgActions {
		if !acl.IsAllowedView(cfg, user, action) {
			continue
		}

		if action.Entity != "" {
			res.Actions = append(res.Actions, buildActionEntities(action.Entity, action)...)
		} else {
			btn := actionCfgToPb(action, user)
			res.Actions = append(res.Actions, btn)
		}
	}

	return res
}

func buildActionEntities(entityTitle string, tpl *config.Action) []*pb.Action {
	ret := make([]*pb.Action, 0)

	entityCount, _ := strconv.Atoi(sv.Get("entities." + entityTitle + ".count"))

	for i := 0; i < entityCount; i++ {
		ret = append(ret, buildEntityAction(tpl, entityTitle, i))
	}

	return ret
}

func buildEntityAction(tpl *config.Action, entityTitle string, entityIndex int) *pb.Action {
	prefix := getEntityPrefix(entityTitle, entityIndex)

	virtualActionId := createPublicID(tpl, prefix)

	publicActionIdToActionMap[virtualActionId] = ActionWithEntity{
		Action:       tpl,
		EntityPrefix: prefix,
	}

	return &pb.Action{
		Id:    virtualActionId,
		Title: sv.ReplaceEntityVars(prefix, tpl.Title),
		Icon:  tpl.Icon,
	}
}

func actionCfgToPb(action *config.Action, user *acl.AuthenticatedUser) *pb.Action {
	virtualActionId := createPublicID(action, "")

	publicActionIdToActionMap[virtualActionId] = ActionWithEntity{
		Action:       action,
		EntityPrefix: "noent",
	}

	btn := pb.Action{
		Id:           virtualActionId,
		Title:        action.Title,
		Icon:         action.Icon,
		CanExec:      acl.IsAllowedExec(cfg, user, action),
		PopupOnStart: action.PopupOnStart,
	}

	for _, cfgArg := range action.Arguments {
		pbArg := pb.ActionArgument{
			Name:         cfgArg.Name,
			Title:        cfgArg.Title,
			Type:         cfgArg.Type,
			Description:  cfgArg.Description,
			DefaultValue: cfgArg.Default,
			Choices:      buildChoices(cfgArg.Choices),
		}

		btn.Arguments = append(btn.Arguments, &pbArg)
	}

	return &btn
}

func buildChoices(choices []config.ActionArgumentChoice) []*pb.ActionArgumentChoice {
	ret := []*pb.ActionArgumentChoice{}

	for _, cfgChoice := range choices {
		pbChoice := pb.ActionArgumentChoice{
			Value: cfgChoice.Value,
			Title: cfgChoice.Title,
		}

		ret = append(ret, &pbChoice)
	}

	return ret
}

func createPublicID(action *config.Action, entityPrefix string) string {
	h := sha256.New()
	h.Write([]byte(action.ID+"."+entityPrefix))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func findActionByPublicID(id string) *config.Action {
	pair, found := publicActionIdToActionMap[id]

	if found {
		log.Infof("findPublic %v, %v", id, pair.Action.ID)
		return pair.Action
	}

	return nil
}
