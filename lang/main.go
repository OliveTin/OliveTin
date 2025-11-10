package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jamesread/golure/pkg/dirs"
	log "github.com/sirupsen/logrus"
)

type LanguageFilev1 struct {
	SchemaVersion int               `json:"schemaVersion"`
	Translations  map[string]string `json:"translations"`
}

type CombinedTranslationsOutput struct {
	Messages map[string]map[string]string `json:"messages"`
}

func main() {
	combinedContent := getCombinedLanguageContent()

	jsonData, err := json.Marshal(combinedContent)

	if err != nil {
		log.Fatalf("Error marshalling combined language content: %v", err)
	}

	err = os.WriteFile("combined_output.json", jsonData, 0644)

	if err != nil {
		log.Fatalf("Error saving combined language content to file: %v", err)
		return
	}

	log.Infof("Combined language content saved to combined_output.json")
}

func getLanguageDir() string {
	dirsToSearch := []string{
		"../lang",
		"../../../../lang/", // Relative to this file, for unit tests
		"/app/lang/",
	}

	dir, _ := dirs.GetFirstExistingDirectory("lang", dirsToSearch)

	return dir
}

func getCombinedLanguageContent() *CombinedTranslationsOutput {
	output := &CombinedTranslationsOutput{
		Messages: make(map[string]map[string]string),
	}

	languageDir := getLanguageDir()

	files, err := os.ReadDir(languageDir)

	if err != nil {
		log.Errorf("Error reading language directory %s: %v", languageDir, err)
		return output
	}

	for _, file := range filterLanguageFiles(files) {
		languageName := strings.Replace(file.Name(), ".yaml", "", 1)

		fullPath := filepath.Join(languageDir, file.Name())
		log.Infof("Loading language file: %s", fullPath)

		content, err := os.ReadFile(fullPath)

		if err != nil {
			log.Errorf("Error reading language file %s: %v", fullPath, err)
			continue
		}

		var yamlData LanguageFilev1

		err = yaml.Unmarshal(content, &yamlData)

		if err != nil {
			log.Errorf("Error reading language file %s: %v", fullPath, err)
			continue
		}

		output.Messages[languageName] = yamlData.Translations
	}

	return output
}

func filterLanguageFiles(files []os.DirEntry) []os.DirEntry {
	ret := make([]os.DirEntry, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		ret = append(ret, file)
	}

	return ret
}

func parseAcceptLanguages(headerLanguage string) []string {
	acceptLanguages := make([]string, 0)

	for _, lang := range strings.Split(headerLanguage, ",") {
		lang = strings.TrimSpace(lang)

		acceptLanguages = append(acceptLanguages, lang)
	}

	return acceptLanguages
}
