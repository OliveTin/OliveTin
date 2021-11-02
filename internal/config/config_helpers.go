package config

func (cfg *Config) FindAction(actionTitle string) *Action {
	for _, action := range cfg.Actions {
		if action.Title == actionTitle {
			return &action
		}
	}

	return nil
}
