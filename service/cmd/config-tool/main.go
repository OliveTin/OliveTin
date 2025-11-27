package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/OliveTin/OliveTin/internal/api"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
)

func printPwd() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	log.Infof("Working directory: %s", pwd)
}

func main() {
	resetPasswords := flag.Bool("passwords", true, "Reset passwords")
	flag.Parse()

	log.Info("Config tool started")

	printPwd()

	k := koanf.New(".")

	configPath, err := filepath.Abs("../config.yaml")
	if err != nil {
		log.Fatalf("Error getting absolute config path: %v", err)
	}

	log.Infof("Loading config from %s", configPath)

	backupOriginalConfig(configPath)

	err = k.Load(file.Provider(configPath), yaml.Parser())

	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	cfg := &config.Config{}

	config.AppendSource(cfg, k, configPath)

	if *resetPasswords {
		resetAllPasswords(k, cfg)
	}

	saveConfig(configPath, k)
}

func backupOriginalConfig(configPath string) {
	originalConfigPath := filepath.Join(filepath.Dir(configPath), "config.original.yaml")

	_, err := os.Stat(originalConfigPath)
	if err == nil {
		log.Infof("Backup already exists at %s, skipping backup to preserve original", originalConfigPath)
		return
	}
	if !os.IsNotExist(err) {
		log.Fatalf("Error checking backup file: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config for backup: %v", err)
	}
	err = os.WriteFile(originalConfigPath, data, 0644)
	if err != nil {
		log.Fatalf("Error writing backup config: %v", err)
	}
	log.Infof("Original config backed up to %s", originalConfigPath)
}

func resetAllPasswords(k *koanf.Koanf, cfg *config.Config) {
	if !cfg.AuthLocalUsers.Enabled || len(cfg.AuthLocalUsers.Users) == 0 {
		log.Info("No local users found, skipping password reset")
		return
	}

	hashedPassword, err := api.CreateHash("password")
	if err != nil {
		log.Fatalf("Error creating password hash: %v", err)
	}

	usersSlice := k.Get("authLocalUsers.users")
	usersSliceTyped, ok := usersSlice.([]interface{})

	if ok && len(usersSliceTyped) > 0 {
		newUsersSlice := make([]interface{}, len(usersSliceTyped))
		for index, userValue := range usersSliceTyped {
			userMap, ok := userValue.(map[string]interface{})
			if !ok {
				log.Warnf("User entry at index %d is not a map, skipping", index)
				newUsersSlice[index] = userValue
				continue
			}

			oldPassword, _ := userMap["password"].(string)
			username, _ := userMap["username"].(string)
			if username == "" {
				username = fmt.Sprintf("user[%d]", index)
			}

			newUserMap := make(map[string]interface{})
			for k, v := range userMap {
				newUserMap[k] = v
			}
			newUserMap["password"] = hashedPassword
			newUsersSlice[index] = newUserMap

			oldHashPreview := oldPassword
			if len(oldPassword) > 20 {
				oldHashPreview = oldPassword[:20]
			}
			log.Infof("Reset password for user '%s' (old hash: %s...)", username, oldHashPreview)
		}
		err = k.Set("authLocalUsers.users", newUsersSlice)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatalf("Error setting users")
		}
	} else {
		for index, user := range cfg.AuthLocalUsers.Users {
			key := "authLocalUsers.users." + strconv.Itoa(index) + ".password"
			err = k.Set(key, hashedPassword)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Fatalf("Error setting user password")
			}

			oldHashPreview := user.Password
			if len(oldHashPreview) > 20 {
				oldHashPreview = oldHashPreview[:20]
			}
			log.Infof("Reset password for user '%s' (old hash: %s...)", user.Username, oldHashPreview)
		}
	}

	log.Infof("Reset %d password(s) to 'password'", len(cfg.AuthLocalUsers.Users))
}

func saveConfig(configPath string, k *koanf.Koanf) {
	out, err := k.Marshal(yaml.Parser())

	if err != nil {
		log.Fatalf("Error marshalling config: %v", err)
	}

	err = os.WriteFile(configPath, out, 0644)
	if err != nil {
		log.Fatalf("Error saving config: %v", err)
	}

	log.Infof("Config saved to %s", configPath)
}
