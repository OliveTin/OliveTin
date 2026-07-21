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

func passwordHashPreview(password string) string {
	if len(password) > 20 {
		return password[:20]
	}

	return password
}

func userDisplayName(username string, index int) string {
	if username == "" {
		return fmt.Sprintf("user[%d]", index)
	}

	return username
}

func copyUserMapWithPassword(userMap map[string]interface{}, hashedPassword string) map[string]interface{} {
	newUserMap := make(map[string]interface{}, len(userMap)+1)
	for key, value := range userMap {
		newUserMap[key] = value
	}
	newUserMap["password"] = hashedPassword

	return newUserMap
}

func resetPasswordInUserMap(userValue interface{}, index int, hashedPassword string) interface{} {
	userMap, ok := userValue.(map[string]interface{})
	if !ok {
		log.Warnf("User entry at index %d is not a map, skipping", index)
		return userValue
	}

	oldPassword, _ := userMap["password"].(string)
	username, _ := userMap["username"].(string)
	log.Infof("Reset password for user '%s' (old hash: %s...)", userDisplayName(username, index), passwordHashPreview(oldPassword))

	return copyUserMapWithPassword(userMap, hashedPassword)
}

func resetPasswordsFromSlice(k *koanf.Koanf, usersSliceTyped []interface{}, hashedPassword string) {
	newUsersSlice := make([]interface{}, len(usersSliceTyped))
	for index, userValue := range usersSliceTyped {
		newUsersSlice[index] = resetPasswordInUserMap(userValue, index, hashedPassword)
	}

	err := k.Set("authLocalUsers.users", newUsersSlice)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Error setting users")
	}
}

func resetPasswordsFromConfig(k *koanf.Koanf, cfg *config.Config, hashedPassword string) {
	for index, user := range cfg.AuthLocalUsers.Users {
		key := "authLocalUsers.users." + strconv.Itoa(index) + ".password"
		err := k.Set(key, hashedPassword)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatalf("Error setting user password")
		}

		log.Infof("Reset password for user '%s' (old hash: %s...)", user.Username, passwordHashPreview(user.Password))
	}
}

func hasLocalUsers(cfg *config.Config) bool {
	return cfg.AuthLocalUsers.Enabled && len(cfg.AuthLocalUsers.Users) > 0
}

func applyPasswordResets(k *koanf.Koanf, cfg *config.Config, hashedPassword string) {
	usersSliceTyped, ok := k.Get("authLocalUsers.users").([]interface{})
	if ok && len(usersSliceTyped) > 0 {
		resetPasswordsFromSlice(k, usersSliceTyped, hashedPassword)
		return
	}

	resetPasswordsFromConfig(k, cfg, hashedPassword)
}

func resetAllPasswords(k *koanf.Koanf, cfg *config.Config) {
	if !hasLocalUsers(cfg) {
		log.Info("No local users found, skipping password reset")
		return
	}

	hashedPassword, err := api.CreateHash("password")
	if err != nil {
		log.Fatalf("Error creating password hash: %v", err)
	}

	applyPasswordResets(k, cfg, hashedPassword)
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
