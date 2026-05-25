package api

import (
	apiv1 "github.com/OliveTin/OliveTin/gen/olivetin/api/v1"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func applyActionExecTriggers(pb *apiv1.Action, cfg *config.Action) {
	if cfg == nil {
		return
	}

	pb.ExecOnStartup = cfg.ExecOnStartup
	pb.ExecOnCron = append([]string(nil), cfg.ExecOnCron...)
	pb.ExecOnFileCreatedInDir = append([]string(nil), cfg.ExecOnFileCreatedInDir...)
	pb.ExecOnFileChangedInDir = append([]string(nil), cfg.ExecOnFileChangedInDir...)
	pb.ExecOnCalendarFile = cfg.ExecOnCalendarFile

	for _, wh := range cfg.ExecOnWebhook {
		pb.ExecOnWebhooks = append(pb.ExecOnWebhooks, &apiv1.ActionWebhookExecHint{
			Template:  wh.Template,
			MatchPath: wh.MatchPath,
		})
	}
}
