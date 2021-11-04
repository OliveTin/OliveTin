package config

func (cfg *Config) FindAction(actionTitle string) *Action {
	for _, action := range cfg.Actions {
		if action.Title == actionTitle {
			return &action
		}
	}

	return nil
}

func (action *Action) FindArg(name string) *ActionArgument {
	for _, arg := range action.Arguments {
		if arg.Name == name {
			return &arg
		}
	}

	return nil
}
