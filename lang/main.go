package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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
	Comment  string                       `json:"_comment"`
	Messages map[string]map[string]string `json:"messages"`
}

func main() {
	combinedContent := getCombinedLanguageContent()

	sortedContent := sortTranslations(combinedContent)

	jsonData, err := json.MarshalIndent(sortedContent, "", "    ")

	if err != nil {
		log.Fatalf("Error marshalling combined language content: %v", err)
	}

	err = os.WriteFile("combined_output.json", jsonData, 0644)

	if err != nil {
		log.Fatalf("Error saving combined language content to file: %v", err)
	}

	log.Infof("Combined language content saved to combined_output.json")
}

// sortTranslations creates a new structure with sorted keys for deterministic output.
func sortTranslations(input *CombinedTranslationsOutput) *CombinedTranslationsOutput {
	sorted := &CombinedTranslationsOutput{
		Comment:  input.Comment,
		Messages: make(map[string]map[string]string),
	}

	// Sort language names
	langNames := make([]string, 0, len(input.Messages))
	for langName := range input.Messages {
		langNames = append(langNames, langName)
	}
	sort.Strings(langNames)

	// For each language, sort the translation keys
	for _, langName := range langNames {
		translations := input.Messages[langName]
		sortedTranslations := make(map[string]string)

		keys := make([]string, 0, len(translations))
		for key := range translations {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sortedTranslations[key] = translations[key]
		}

		sorted.Messages[langName] = sortedTranslations
	}

	return sorted
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
		Comment:  "This file is generated. Please re-generate this file using 'make' when you update a translation.",
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

	validateTranslations(output)

	return output
}

// getReferenceKeys returns the keys from the "en" translation as the reference set.
func getReferenceKeys(messages map[string]map[string]string) map[string]bool {
	enTranslations, exists := messages["en"]
	if !exists {
		return nil
	}

	referenceKeys := make(map[string]bool, len(enTranslations))
	for key := range enTranslations {
		referenceKeys[key] = true
	}
	return referenceKeys
}

// findMissingKeys returns the keys that are in referenceKeys but not in translations.
func findMissingKeys(referenceKeys map[string]bool, translations map[string]string) []string {
	missing := make([]string, 0)
	for key := range referenceKeys {
		if _, exists := translations[key]; !exists {
			missing = append(missing, key)
		}
	}
	return missing
}

// findExtraKeys returns the keys that are in translations but not in referenceKeys.
func findExtraKeys(referenceKeys map[string]bool, translations map[string]string) []string {
	extra := make([]string, 0)
	for key := range translations {
		if !referenceKeys[key] {
			extra = append(extra, key)
		}
	}
	return extra
}

// validateTranslations checks all translations against the "en" reference and prints warnings for missing and extra keys.
func validateTranslations(output *CombinedTranslationsOutput) {
	referenceKeys := getReferenceKeys(output.Messages)
	if referenceKeys == nil {
		log.Warnf("No 'en' translation found, skipping validation")
		return
	}

	for langName, translations := range output.Messages {
		if langName == "en" {
			continue
		}

		missing := findMissingKeys(referenceKeys, translations)
		if len(missing) > 0 {
			log.Warnf("Translation '%s' is missing %d key(s): %v", langName, len(missing), missing)
		}

		extra := findExtraKeys(referenceKeys, translations)
		if len(extra) > 0 {
			log.Warnf("Translation '%s' has %d extra key(s) not in 'en': %v", langName, len(extra), extra)
		}
	}
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
