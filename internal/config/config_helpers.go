package config

// FindAction will return a action if there is a match on Title
func (cfg *Config) FindAction(actionTitle string) *Action {
	for _, action := range cfg.Actions {
		if action.Title == actionTitle {
			return &action
		}
	}

	return nil
}

// FindArg will return an arg if there is a match on Name
func (action *Action) FindArg(name string) *ActionArgument {
	for _, arg := range action.Arguments {
		if arg.Name == name {
			return &arg
		}
	}

	return nil
}
