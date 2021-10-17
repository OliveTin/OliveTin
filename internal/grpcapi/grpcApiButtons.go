package grpcapi

import (
	"fmt"
	"crypto/md5"
	pb "github.com/jamesread/OliveTin/gen/grpc"
	config "github.com/jamesread/OliveTin/internal/config"
	acl "github.com/jamesread/OliveTin/internal/acl"
)

func buildButton(action config.ActionButton, user *acl.User) (*pb.ActionButton) {
	btn := pb.ActionButton{
		Id:      fmt.Sprintf("%x", md5.Sum([]byte(action.Title))),
		Title:   action.Title,
		Icon:    lookupHTMLIcon(action.Icon),
		CanExec: acl.IsAllowedExec(cfg, user, &action),
	}

	for _, cfgArg := range action.Arguments {
		pbArg := pb.ActionArgument {
			Name: cfgArg.Name,
			Label: cfgArg.Label,
			Type: cfgArg.Type,
			DefaultValue: cfgArg.Default,
			Choices: buildChoices(cfgArg.Choices),
		}

		btn.Arguments = append(btn.Arguments, &pbArg)
	}

	return &btn
}

func buildChoices(choices []config.ActionArgumentChoice) ([]*pb.ActionArgumentChoice) {
	ret := []*pb.ActionArgumentChoice{}

	for _, cfgChoice := range choices {
		pbChoice := pb.ActionArgumentChoice {
			Value: cfgChoice.Value,
			Label: cfgChoice.Label,
		}

		ret = append(ret, &pbChoice);
	}

	return ret;
}
