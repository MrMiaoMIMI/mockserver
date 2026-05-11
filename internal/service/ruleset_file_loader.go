package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mockserver/internal/model/bo"
)

func LoadRuleSetFiles(ctx context.Context, ruleSetService RuleSetService, path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	paths, err := expandRuleSetPaths(path)
	if err != nil {
		return err
	}
	for _, item := range paths {
		if err := loadSingleRuleSetFile(ctx, ruleSetService, item); err != nil {
			return err
		}
	}
	return nil
}

func expandRuleSetPaths(raw string) ([]string, error) {
	entries := strings.Split(raw, ",")
	paths := make([]string, 0)
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		info, err := os.Stat(entry)
		if err != nil {
			return nil, fmt.Errorf("stat ruleset path %s: %w", entry, err)
		}
		if !info.IsDir() {
			paths = append(paths, entry)
			continue
		}

		dirEntries, err := os.ReadDir(entry)
		if err != nil {
			return nil, fmt.Errorf("read ruleset directory %s: %w", entry, err)
		}

		jsonFiles := make([]string, 0, len(dirEntries))
		for _, dirEntry := range dirEntries {
			if dirEntry.IsDir() {
				continue
			}
			if strings.ToLower(filepath.Ext(dirEntry.Name())) != ".json" {
				continue
			}
			jsonFiles = append(jsonFiles, filepath.Join(entry, dirEntry.Name()))
		}
		sort.Strings(jsonFiles)
		if len(jsonFiles) == 0 {
			return nil, fmt.Errorf("no ruleset json files found in directory %s", entry)
		}
		paths = append(paths, jsonFiles...)
	}

	if len(paths) == 0 {
		return nil, nil
	}
	return paths, nil
}

func loadSingleRuleSetFile(ctx context.Context, ruleSetService RuleSetService, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ruleset file %s: %w", path, err)
	}

	var ruleSet bo.RuleSet
	if err := json.Unmarshal(content, &ruleSet); err != nil {
		return fmt.Errorf("decode ruleset file %s: %w", path, err)
	}

	saved, err := ruleSetService.UpsertDraft(ctx, ruleSet)
	if err != nil {
		return fmt.Errorf("upsert draft from file %s: %w", path, err)
	}

	validation, err := ruleSetService.ValidateDraft(ctx, saved.ID)
	if err != nil {
		return fmt.Errorf("validate ruleset %s from file %s: %w", saved.ID, path, err)
	}
	if !validation.Valid {
		return fmt.Errorf("ruleset %s from file %s is invalid: %+v", saved.ID, path, validation.Issues)
	}

	if _, err := ruleSetService.Publish(ctx, saved.ID, bo.AuditInfo{
		Operator: "bootstrap",
		Reason:   "startup ruleset file load",
	}); err != nil {
		return fmt.Errorf("publish ruleset %s from file %s: %w", saved.ID, path, err)
	}

	return nil
}
