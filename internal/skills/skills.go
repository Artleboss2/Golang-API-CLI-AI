package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Prompt      string
}

func cliDir() string {
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe)
	}
	self, err := exec.LookPath(os.Args[0])
	if err == nil {
		abs, err := filepath.Abs(self)
		if err == nil {
			return filepath.Dir(abs)
		}
	}
	abs, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return abs
}

func Dir() (string, error) {
	return filepath.Join(cliDir(), "skills"), nil
}

func Load(name string) (*Skill, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	candidates := []string{
		filepath.Join(dir, name+".md"),
		filepath.Join(dir, name+".txt"),
		filepath.Join(dir, name),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return parseSkill(name, string(data)), nil
		}
	}

	return nil, fmt.Errorf("skill %q introuvable dans %s", name, dir)
}

func List() ([]Skill, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le dossier skills : %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("lecture du dossier skills échouée : %w", err)
	}

	var skillList []Skill
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".md" && ext != ".txt" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ext)
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		skillList = append(skillList, *parseSkill(name, string(data)))
	}
	return skillList, nil
}

func Save(name, content string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le dossier skills : %w", err)
	}
	return os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0644)
}

func Delete(name string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	candidates := []string{
		filepath.Join(dir, name+".md"),
		filepath.Join(dir, name+".txt"),
	}
	for _, p := range candidates {
		if err := os.Remove(p); err == nil {
			return nil
		}
	}
	return fmt.Errorf("skill %q introuvable", name)
}

func BuildSystemAddendum(names []string) string {
	if len(names) == 0 {
		return ""
	}
	var parts []string
	for _, name := range names {
		s, err := Load(name)
		if err != nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("[Skill : %s]\n%s", s.Name, s.Prompt))
	}
	if len(parts) == 0 {
		return ""
	}
	return "\n\n" + strings.Join(parts, "\n\n")
}

func parseSkill(name, raw string) *Skill {
	skill := &Skill{Name: name}
	lines := strings.Split(raw, "\n")

	var promptLines []string
	inPrompt := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && skill.Description == "" {
			skill.Description = strings.TrimPrefix(trimmed, "# ")
			continue
		}
		if strings.EqualFold(trimmed, "## prompt") || strings.EqualFold(trimmed, "---") {
			inPrompt = true
			continue
		}
		if inPrompt {
			promptLines = append(promptLines, line)
		}
	}

	if len(promptLines) > 0 {
		skill.Prompt = strings.TrimSpace(strings.Join(promptLines, "\n"))
	} else {
		skill.Prompt = strings.TrimSpace(raw)
	}

	if skill.Description == "" {
		skill.Description = name
	}

	return skill
}
