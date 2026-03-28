package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	nimapi "github.com/user/nvidia-nim-cli/internal/api"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/filewriter"
	"github.com/user/nvidia-nim-cli/internal/skills"
	"github.com/user/nvidia-nim-cli/internal/titler"
	"github.com/user/nvidia-nim-cli/internal/ui"
	"github.com/user/nvidia-nim-cli/pkg/models"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Démarrer une session de chat interactive avec un modèle NIM",
	Long:  `Lance une session de chat interactive avec historique de conversation.`,
	RunE:  runChat,
}

var (
	chatModel       string
	chatSystem      string
	chatStream      bool
	chatTemp        float64
	chatNoAutoTitle bool
)

func init() {
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "", "Modèle NIM à utiliser")
	chatCmd.Flags().StringVarP(&chatSystem, "system", "s", "", "Message système personnalisé")
	chatCmd.Flags().BoolVar(&chatStream, "stream", true, "Activer le streaming des réponses")
	chatCmd.Flags().Float64VarP(&chatTemp, "temperature", "t", -1, "Température (0.0–1.0)")
	chatCmd.Flags().BoolVar(&chatNoAutoTitle, "no-title", false, "Désactiver le titre automatique")
}

type ChatSession struct {
	client       *nimapi.Client
	model        string
	history      []nimapi.Message
	stream       bool
	temp         float64
	msgCount     int
	started      time.Time
	title        string
	sessionDir   string
	titleSet     bool
	activeSkills []string
}

func runChat(cmd *cobra.Command, args []string) error {
	if !config.HasAPIKey() {
		ui.PrintError("Aucune clé API configurée.")
		fmt.Println(ui.InfoStyle.Render("Configurez votre clé avec : ") + ui.CommandStyle.Render("nim auth"))
		return nil
	}

	model := chatModel
	if model == "" {
		model = config.GetModel()
	}

	temperature := chatTemp
	if temperature < 0 {
		temperature = config.GetTemperature()
	}

	selectedSkills, err := promptSkillSelection()
	if err != nil {
		return err
	}

	session := &ChatSession{
		client:       nimapi.NewClient(config.GetAPIKey()),
		model:        model,
		stream:       chatStream,
		temp:         temperature,
		started:      time.Now(),
		title:        titler.Fallback(),
		activeSkills: selectedSkills,
	}

	session.sessionDir = filewriter.SessionDir(session.title)

	systemMsg := chatSystem
	if systemMsg == "" {
		systemMsg = buildSystemPrompt(session.activeSkills)
	}

	session.history = []nimapi.Message{
		{Role: "system", Content: systemMsg},
	}

	return session.run()
}

func promptSkillSelection() ([]string, error) {
	allSkills, err := skills.List()
	if err != nil || len(allSkills) == 0 {
		return nil, nil
	}

	items := make([]ui.MenuItem, len(allSkills))
	for i, sk := range allSkills {
		desc := sk.Meta.Description
		if sk.Meta.Category != "" {
			desc = "[" + sk.Meta.Category + "] " + desc
		}
		items[i] = ui.MenuItem{
			ID:    sk.Meta.Name,
			Label: sk.Meta.Name,
			Desc:  desc,
		}
	}

	selected, escaped := ui.RunMultiSelect("Sélectionner les skills à activer (optionnel)", items)
	if escaped {
		return nil, nil
	}
	return selected, nil
}

func buildSystemPrompt(activeSkills []string) string {
	base := `Tu es un assistant IA intelligent et concis. Réponds en français sauf si l'utilisateur parle dans une autre langue. Sois précis, utile et professionnel.

Tu peux créer des fichiers directement sur le PC de l'utilisateur avec ce format exact :
<file:nom_du_fichier.ext>
contenu du fichier
</file>

Règles de création de fichiers :
1. Utilise ce format UNIQUEMENT quand l'utilisateur demande de créer, écrire ou générer un fichier.
2. Tu peux inclure des sous-dossiers : <file:src/main.go>
3. Tu peux créer plusieurs fichiers dans une seule réponse.
4. Le contenu entre les balises est écrit tel quel sur le disque.
5. Après les blocs <file:...>, explique brièvement ce que tu as créé.`

	addendum := skills.BuildSystemAddendum(activeSkills)
	return base + addendum
}

func (s *ChatSession) run() error {
	fmt.Println()
	fmt.Println(ui.SmallBanner(models.GetDisplayName(s.model)))
	fmt.Println()
	s.printWelcome()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for {
		fmt.Print("\n" + ui.PromptStyle.Render("Vous ▶ "))

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "/" {
			chosen, escaped := s.showCommandMenu()
			if escaped || chosen == "" {
				continue
			}
			if s.handleCommand(chosen) {
				break
			}
			continue
		}

		if strings.HasPrefix(input, "/") {
			if s.handleCommand(input) {
				break
			}
			continue
		}

		if err := s.sendMessage(input); err != nil {
			ui.PrintError(err.Error())
		}
	}

	s.printSessionSummary()
	return nil
}

func (s *ChatSession) showCommandMenu() (string, bool) {
	items := []ui.CommandMenuItem{
		{Key: "/skill", Label: "Gérer les skills", Desc: "Ajouter ou retirer des skills"},
		{Key: "/model", Label: "Changer de modèle", Desc: "Voir ou changer le modèle actif"},
		{Key: "/stream", Label: "Toggle streaming", Desc: "Activer / désactiver le streaming"},
		{Key: "/clear", Label: "Effacer l'historique", Desc: "Recommencer la conversation"},
		{Key: "/history", Label: "Voir l'historique", Desc: "Afficher les messages de la session"},
		{Key: "/save", Label: "Sauvegarder", Desc: "Enregistrer l'historique dans un fichier"},
		{Key: "/title", Label: "Titre & dossier", Desc: "Afficher le titre et le dossier de session"},
		{Key: "/system", Label: "Message système", Desc: "Voir ou modifier le prompt système"},
		{Key: "/quit", Label: "Quitter", Desc: "Terminer la session"},
	}
	return ui.RunCommandMenu(items)
}

func (s *ChatSession) sendMessage(content string) error {
	s.history = append(s.history, nimapi.Message{Role: "user", Content: content})
	s.msgCount++

	if !s.titleSet && !chatNoAutoTitle && s.msgCount == 1 {
		go s.resolveTitle(content)
	}

	req := &nimapi.ChatRequest{
		Model:       s.model,
		Messages:    s.history,
		MaxTokens:   config.GetMaxTokens(),
		Temperature: s.temp,
		Stream:      s.stream,
	}

	fmt.Println()
	fmt.Print(ui.AIPrefixStyle.Render("◆ " + models.GetDisplayName(s.model) + " ▶ "))
	fmt.Println()
	fmt.Println()

	var fullResponse string
	var err error

	if s.stream {
		fullResponse, err = s.streamResponse(req)
	} else {
		fullResponse, err = s.standardResponse(req)
	}
	if err != nil {
		return err
	}

	cleanResponse, createdFiles := filewriter.ExtractAndWrite(fullResponse, s.sessionDir)

	if !s.stream {
		fmt.Println(ui.AIMessageStyle.Render(cleanResponse))
	}

	printCreatedFiles(createdFiles)

	if cleanResponse != "" {
		s.history = append(s.history, nimapi.Message{Role: "assistant", Content: cleanResponse})
	}

	fmt.Printf("\n%s %s\n",
		ui.TimestampStyle.Render(ui.FormatTimestamp(time.Now())),
		ui.InfoStyle.Render(fmt.Sprintf("│ Message #%d │ Dossier : %s", s.msgCount, s.sessionDir)),
	)

	return nil
}

func (s *ChatSession) streamResponse(req *nimapi.ChatRequest) (string, error) {
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go s.client.StreamComplete(req, tokenChan, errChan)

	var sb strings.Builder

loop:
	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				break loop
			}
			fmt.Print(ui.AIMessageStyle.Render(token))
			sb.WriteString(token)
		case err := <-errChan:
			if err != nil {
				fmt.Println()
				return "", fmt.Errorf("erreur streaming : %w", err)
			}
		}
	}
	fmt.Println()
	return sb.String(), nil
}

func (s *ChatSession) standardResponse(req *nimapi.ChatRequest) (string, error) {
	spinner := ui.NewSpinner("Génération en cours...")
	spinner.Start()

	resp, err := s.client.Complete(req)
	if err != nil {
		spinner.StopWithError("")
		return "", err
	}
	spinner.Stop("")

	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}

func (s *ChatSession) resolveTitle(firstMsg string) {
	title, err := titler.Generate(config.GetAPIKey(), s.model, firstMsg)
	if err != nil || strings.TrimSpace(title) == "" {
		return
	}
	s.title = title
	s.sessionDir = filewriter.SessionDir(title)
	s.titleSet = true
}

func (s *ChatSession) handleCommand(cmd string) bool {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}
	command := strings.ToLower(parts[0])

	switch command {
	case "/quit", "/exit", "/q":
		return true

	case "/clear", "/cls":
		system := s.history[0]
		s.history = []nimapi.Message{system}
		s.msgCount = 0
		s.titleSet = false
		ui.PrintSuccess("Historique effacé.")

	case "/history":
		s.printHistory()

	case "/model":
		if len(parts) > 1 {
			s.model = parts[1]
			ui.PrintSuccess("Modèle changé : " + models.GetDisplayName(s.model))
		} else {
			fmt.Println(ui.InfoStyle.Render("Modèle actuel : ") + ui.ModelStyle.Render(models.GetDisplayName(s.model)))
		}

	case "/stream":
		s.stream = !s.stream
		status := "désactivé"
		if s.stream {
			status = "activé"
		}
		ui.PrintSuccess("Streaming " + status)

	case "/save":
		if err := s.saveHistory(); err != nil {
			ui.PrintError("Sauvegarde échouée : " + err.Error())
		} else {
			ui.PrintSuccess("Historique sauvegardé dans " + s.sessionDir)
		}

	case "/system":
		if len(parts) > 1 {
			s.history[0] = nimapi.Message{Role: "system", Content: strings.Join(parts[1:], " ")}
			ui.PrintSuccess("Message système mis à jour.")
		} else {
			fmt.Println(ui.InfoStyle.Render("Système actuel : ") + s.history[0].Content)
		}

	case "/skill":
		s.handleSkillCommand(parts)

	case "/title":
		fmt.Println(ui.InfoStyle.Render("Titre   : ") + ui.ModelStyle.Render(s.title))
		fmt.Println(ui.InfoStyle.Render("Dossier : ") + s.sessionDir)
	}

	return false
}

func (s *ChatSession) handleSkillCommand(parts []string) {
	allSkills, err := skills.List()
	if err != nil {
		ui.PrintError(err.Error())
		return
	}

	if len(allSkills) == 0 {
		base, _ := skills.BaseDir()
		ui.PrintWarning("Aucun skill disponible dans " + base)
		return
	}

	items := make([]ui.MenuItem, len(allSkills))
	for i, sk := range allSkills {
		desc := sk.Meta.Description
		if sk.Meta.Category != "" {
			desc = "[" + sk.Meta.Category + "] " + desc
		}
		items[i] = ui.MenuItem{
			ID:    sk.Meta.Name,
			Label: sk.Meta.Name,
			Desc:  desc,
		}
	}

	selected, escaped := ui.RunMultiSelect("Sélectionner les skills actifs", items)
	if escaped {
		return
	}

	s.activeSkills = selected
	s.history[0] = nimapi.Message{
		Role:    "system",
		Content: buildSystemPrompt(s.activeSkills),
	}

	if len(selected) == 0 {
		ui.PrintSuccess("Aucun skill actif.")
	} else {
		ui.PrintSuccess("Skills actifs : " + strings.Join(selected, ", "))
	}
}

func (s *ChatSession) printWelcome() {
	line := ui.InfoStyle.Render("Tapez ") +
		ui.CommandStyle.Render("/") +
		ui.InfoStyle.Render(" pour ouvrir le menu des commandes • ") +
		ui.CommandStyle.Render("/quit") +
		ui.InfoStyle.Render(" pour quitter")

	if len(s.activeSkills) > 0 {
		line += "\n" + ui.InfoStyle.Render("Skills actifs : ") + ui.SuccessStyle.Render(strings.Join(s.activeSkills, ", "))
	}
	fmt.Println(line)
}

func (s *ChatSession) printHistory() {
	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  Historique de la session"))
	fmt.Println(ui.Divider(60))

	for i, msg := range s.history {
		if msg.Role == "system" {
			continue
		}
		var prefix string
		if msg.Role == "user" {
			prefix = ui.UserPrefixStyle.Render(fmt.Sprintf("[%d] Vous", i))
		} else {
			prefix = ui.AIPrefixStyle.Render(fmt.Sprintf("[%d] IA  ", i))
		}
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("%s : %s\n\n", prefix, ui.InfoStyle.Render(content))
	}
	fmt.Println(ui.Divider(60))
}

func (s *ChatSession) saveHistory() error {
	if err := os.MkdirAll(s.sessionDir, 0755); err != nil {
		return fmt.Errorf("impossible de créer le dossier : %w", err)
	}

	path := filepath.Join(s.sessionDir, fmt.Sprintf("historique_%s.txt", time.Now().Format("20060102_150405")))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Session NVIDIA NIM — %s\n", s.title)
	fmt.Fprintf(f, "Démarré : %s\n", s.started.Format("02/01/2006 15:04:05"))
	fmt.Fprintf(f, "Modèle : %s\n", s.model)
	if len(s.activeSkills) > 0 {
		fmt.Fprintf(f, "Skills : %s\n", strings.Join(s.activeSkills, ", "))
	}
	fmt.Fprintf(f, "Messages : %d\n", s.msgCount)
	fmt.Fprintf(f, "%s\n\n", strings.Repeat("─", 60))

	for _, msg := range s.history {
		if msg.Role == "system" {
			continue
		}
		role := "Vous"
		if msg.Role == "assistant" {
			role = "IA"
		}
		fmt.Fprintf(f, "[%s]\n%s\n\n", role, msg.Content)
	}
	return nil
}

func (s *ChatSession) printSessionSummary() {
	duration := time.Since(s.started).Round(time.Second)
	fmt.Println()
	fmt.Println(ui.Divider(60))
	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf(
		"  Session terminée │ %d messages │ Durée : %s",
		s.msgCount, duration,
	)))
	if s.msgCount > 0 {
		fmt.Println(ui.InfoStyle.Render("  Titre    : " + s.title))
		fmt.Println(ui.InfoStyle.Render("  Dossier  : " + s.sessionDir))
	}
	fmt.Println(ui.Divider(60))
	fmt.Println()
}
