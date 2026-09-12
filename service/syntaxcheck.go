package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/executor"
)

func runSyntaxCheck() {
	printSyntaxCheckSearchResults()
	printLoadedConfigFiles()

	if syntaxCheckLoadError != nil {
		fmt.Printf("Configuration load failed: %v\n", syntaxCheckLoadError)
		fmt.Println("0 configuration issues found")
		os.Exit(1)
	}

	executor.DefaultExecutor(cfg).RebuildActionMap()
	printConfigIssues(configissues.List())

	issueCount := configissues.Count()
	fmt.Printf("%d configuration issues found\n", issueCount)

	if issueCount > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func printSyntaxCheckSearchResults() {
	for _, line := range syntaxCheckSearchResults {
		fmt.Println(line)
	}
}

func printLoadedConfigFiles() {
	files := config.LoadedSources()
	if len(files) == 0 {
		fmt.Println("No config files were read")
		return
	}

	fmt.Println("Config files read:")
	for _, path := range files {
		fmt.Printf("  %s\n", path)
	}
}

func printConfigIssues(issues []configissues.Issue) {
	if len(issues) == 0 {
		return
	}

	fmt.Println("Configuration issues:")
	for _, issue := range issues {
		fmt.Printf("  %s\n", formatConfigIssue(issue))
	}
}

func formatConfigIssue(issue configissues.Issue) string {
	parts := []string{issue.Severity, issue.Code}

	if issue.ConfigFile != "" {
		parts = append(parts, "file="+issue.ConfigFile)
	}
	if issue.ActionTitle != "" {
		parts = append(parts, "action="+issue.ActionTitle)
	}
	if issue.ArgumentName != "" {
		parts = append(parts, "argument="+issue.ArgumentName)
	}

	return strings.Join(parts, " ") + ": " + issue.Message
}
