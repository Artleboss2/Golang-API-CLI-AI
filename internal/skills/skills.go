package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Meta struct {
	Name        string
	Description string
	Category    string
	Version     string
	Author      string
	Tags        []string
	Prompt      string
}

type Skill struct {
	Meta
	Readme string
	Dir    string
}

func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("répertoire home introuvable : %w", err)
	}
	return filepath.Join(home, "NIM_CLI", "skills"), nil
}

func Load(name string) (*Skill, error) {
	base, err := BaseDir()
	if err != nil {
		return nil, err
	}

	skillDir := filepath.Join(base, name)
	info, err := os.Stat(skillDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("skill %q introuvable dans %s", name, base)
	}

	yamlPath := filepath.Join(skillDir, "skill.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("skill.yaml manquant dans %q : %w", skillDir, err)
	}

	meta := parseYAML(string(data))
	if meta.Name == "" {
		meta.Name = name
	}

	skill := &Skill{Meta: meta, Dir: skillDir}

	if readme, err := os.ReadFile(filepath.Join(skillDir, "README.md")); err == nil {
		skill.Readme = strings.TrimSpace(string(readme))
	}

	return skill, nil
}

func List() ([]Skill, error) {
	base, err := BaseDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(base, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le dossier skills : %w", err)
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("lecture du dossier skills échouée : %w", err)
	}

	var list []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sk, err := Load(e.Name())
		if err != nil {
			continue
		}
		list = append(list, *sk)
	}
	return list, nil
}

func Create(name string, meta Meta) error {
	base, err := BaseDir()
	if err != nil {
		return err
	}

	skillDir := filepath.Join(base, name)
	if err := os.MkdirAll(filepath.Join(skillDir, "assets"), 0755); err != nil {
		return fmt.Errorf("impossible de créer le skill : %w", err)
	}

	if meta.Name == "" {
		meta.Name = name
	}

	yaml := buildYAML(meta)
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(yaml), 0644); err != nil {
		return err
	}

	readme := fmt.Sprintf("# %s\n\n%s\n\n## Usage\n\nDécrivez ici comment utiliser ce skill.\n\n## Assets\n\nPlacez vos fichiers de référence dans le dossier `assets/`.\n", meta.Name, meta.Description)
	return os.WriteFile(filepath.Join(skillDir, "README.md"), []byte(readme), 0644)
}

func Delete(name string) error {
	base, err := BaseDir()
	if err != nil {
		return err
	}
	skillDir := filepath.Join(base, name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q introuvable", name)
	}
	return os.RemoveAll(skillDir)
}

func BuildSystemBlock(sk *Skill) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[Skill actif : %s]\n", sk.Meta.Name))
	if sk.Meta.Description != "" {
		sb.WriteString(fmt.Sprintf("Description : %s\n", sk.Meta.Description))
	}
	if sk.Meta.Prompt != "" {
		sb.WriteString("\nInstructions :\n")
		sb.WriteString(sk.Meta.Prompt)
		sb.WriteString("\n")
	}
	if sk.Readme != "" {
		sb.WriteString("\nContexte détaillé :\n")
		sb.WriteString(sk.Readme)
		sb.WriteString("\n")
	}
	return sb.String()
}

func BuildSystemAddendum(names []string) string {
	if len(names) == 0 {
		return ""
	}
	var blocks []string
	for _, name := range names {
		sk, err := Load(name)
		if err != nil {
			continue
		}
		blocks = append(blocks, BuildSystemBlock(sk))
	}
	if len(blocks) == 0 {
		return ""
	}
	return "\n\n" + strings.Join(blocks, "\n---\n")
}

func parseYAML(raw string) Meta {
	meta := Meta{}
	lines := strings.Split(raw, "\n")
	inPrompt := false
	var promptLines []string

	for _, line := range lines {
		if inPrompt {
			promptLines = append(promptLines, line)
			continue
		}
		if !strings.Contains(line, ":") {
			continue
		}
		idx := strings.Index(line, ":")
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, "\"'")

		switch strings.ToLower(key) {
		case "name":
			meta.Name = val
		case "description":
			meta.Description = val
		case "category":
			meta.Category = val
		case "version":
			meta.Version = val
		case "author":
			meta.Author = val
		case "tags":
			for _, t := range strings.Split(val, ",") {
				tag := strings.TrimSpace(t)
				if tag != "" {
					meta.Tags = append(meta.Tags, tag)
				}
			}
		case "prompt":
			if strings.HasPrefix(val, "|") || val == "" {
				inPrompt = true
			} else {
				meta.Prompt = val
			}
		}
	}

	if len(promptLines) > 0 {
		dedented := dedentLines(promptLines)
		meta.Prompt = strings.TrimSpace(strings.Join(dedented, "\n"))
	}

	return meta
}

func dedentLines(lines []string) []string {
	minIndent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		indent := len(l) - len(strings.TrimLeft(l, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return lines
	}
	result := make([]string, len(lines))
	for i, l := range lines {
		if len(l) >= minIndent {
			result[i] = l[minIndent:]
		} else {
			result[i] = l
		}
	}
	return result
}

func buildYAML(meta Meta) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("name: %q\n", meta.Name))
	sb.WriteString(fmt.Sprintf("description: %q\n", meta.Description))
	sb.WriteString(fmt.Sprintf("category: %q\n", meta.Category))
	sb.WriteString(fmt.Sprintf("version: %q\n", meta.Version))
	sb.WriteString(fmt.Sprintf("author: %q\n", meta.Author))
	if len(meta.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: %q\n", strings.Join(meta.Tags, ", ")))
	}
	if meta.Prompt != "" {
		sb.WriteString("prompt: |\n")
		for _, line := range strings.Split(meta.Prompt, "\n") {
			sb.WriteString("  " + line + "\n")
		}
	}
	return sb.String()
}
