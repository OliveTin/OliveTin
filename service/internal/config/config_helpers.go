package config

// FindAction will return a action if there is a match on Title
func (cfg *Config) findAction(actionTitle string) *Action {
	for _, action := range cfg.Actions {
		if action.Title == actionTitle {
			return action
		}
	}

	return nil
}

// FindArg will return an arg if there is a match on Name
func (action *Action) FindArg(name string) *ActionArgument {
	if name == "stdout" || name == "exitCode" {
		return &ActionArgument{
			Name: name,
			Type: "very_dangerous_raw_string",
		}
	}

	return action.findArg(name)
}

func (action *Action) findArg(name string) *ActionArgument {
	for _, arg := range action.Arguments {
		if arg.Name == name {
			return &arg
		}
	}

	return nil
}

func (cfg *Config) FindAcl(aclTitle string) *AccessControlList {
	for _, acl := range cfg.AccessControlLists {
		if acl.Name == aclTitle {
			return acl
		}
	}

	return nil
}

func (cfg *Config) FindUserByUsername(searchUsername string) *LocalUser {
	for _, user := range cfg.AuthLocalUsers.Users {
		if user.Username == searchUsername {
			return user
		}
	}

	return nil
}

func (cfg *Config) SetDir(dir string) {
	cfg.usedConfigDir = dir
}

func (cfg *Config) GetDir() string {
	return cfg.usedConfigDir
}
