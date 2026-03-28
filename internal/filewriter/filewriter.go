package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var fileBlockRegex = regexp.MustCompile(`(?s)<file:([^>\s]+)>\n?(.*?)\n?</file>`)

type FileResult struct {
	Name    string
	AbsPath string
	Err     error
}

func baseDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "C:\\Users\\Default"
	}
	return filepath.Join(home, "NIM_CLI")
}

func FilesDir() string {
	return filepath.Join(baseDir(), "files")
}

func SkillsDir() string {
	return filepath.Join(baseDir(), "skills")
}

func SessionDir(title string) string {
	return filepath.Join(FilesDir(), sanitizeTitle(title))
}

func ExtractAndWrite(response, sessionDir string) (string, []FileResult) {
	matches := fileBlockRegex.FindAllStringSubmatchIndex(response, -1)
	if len(matches) == 0 {
		return response, nil
	}

	var results []FileResult

	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		rawName := strings.TrimSpace(response[m[2]:m[3]])
		content := response[m[4]:m[5]]

		result := writeFile(rawName, content, sessionDir)
		results = append([]FileResult{result}, results...)
		response = response[:m[0]] + response[m[1]:]
	}

	return strings.TrimSpace(response), results
}

func writeFile(rawName, content, sessionDir string) FileResult {
	result := FileResult{Name: rawName}

	clean := filepath.Clean(rawName)
	if strings.HasPrefix(clean, "..") || clean == "." {
		result.Err = fmt.Errorf("nom de fichier invalide : %q", rawName)
		return result
	}

	target := filepath.Join(sessionDir, clean)
	dir := filepath.Dir(target)

	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Err = fmt.Errorf("impossible de créer le dossier %q : %w", dir, err)
		return result
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		result.Err = fmt.Errorf("chemin absolu introuvable : %w", err)
		return result
	}
	result.AbsPath = abs

	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		result.Err = fmt.Errorf("écriture échouée : %w", err)
		return result
	}

	return result
}

func sanitizeTitle(title string) string {
	title = strings.TrimSpace(title)
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "", "\"", "", "<", "", ">", "", "|", "-",
	)
	title = replacer.Replace(title)
	title = strings.Join(strings.Fields(title), "_")
	if len(title) > 60 {
		title = title[:60]
	}
	if title == "" {
		title = "session"
	}
	return title
}
