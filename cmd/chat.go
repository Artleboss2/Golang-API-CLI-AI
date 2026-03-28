package cmd

import (
	"bufio"
	"fmt"
	"os"
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
	Long: `Lance une session de chat interactive avec historique de conversation.

Le modèle se souvient de tout le contexte de la session.
Tapez '/help' dans le chat pour voir les commandes disponibles.`,
	RunE: runChat,
}

var (
	chatModel       string
	chatSystem      string
	chatStream      bool
	chatTemp        float64
	chatSkills      []string
	chatNoAutoTitle bool
)

func init() {
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "", "Modèle NIM à utiliser")
	chatCmd.Flags().StringVarP(&chatSystem, "system", "s", "", "Message système personnalisé")
	chatCmd.Flags().BoolVar(&chatStream, "stream", true, "Activer le streaming des réponses")
	chatCmd.Flags().Float64VarP(&chatTemp, "temperature", "t", -1, "Température (0.0–1.0)")
	chatCmd.Flags().StringSliceVarP(&chatSkills, "skill", "k", nil, "Skills à activer (ex: -k python -k tests)")
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

	session := &ChatSession{
		client:       nimapi.NewClient(config.GetAPIKey()),
		model:        model,
		stream:       chatStream,
		temp:         temperature,
		started:      time.Now(),
		title:        titler.Fallback(),
		activeSkills: chatSkills,
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

func buildSystemPrompt(activeSkills []string) string {
	base := `Tu es un assistant IA intelligent et concis. Réponds en français sauf si l'utilisateur parle dans une autre langue. Sois précis, utile et professionnel.

Tu peux créer des fichiers directement sur le PC de l'utilisateur en utilisant ce format exact :
<file:nom_du_fichier.ext>
contenu du fichier
</file>

Règles de création de fichiers :
1. Utilise ce format UNIQUEMENT quand l'utilisateur demande de créer, écrire ou générer un fichier.
2. Tu peux inclure des sous-dossiers dans le nom : <file:src/main.go>
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
	printChatWelcome(s.activeSkills)

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

func (s *ChatSession) sendMessage(content string) error {
	s.history = append(s.history, nimapi.Message{
		Role:    "user",
		Content: content,
	})
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
		s.history = append(s.history, nimapi.Message{
			Role:    "assistant",
			Content: cleanResponse,
		})
	}

	fmt.Printf("\n%s %s\n",
		ui.TimestampStyle.Render(ui.FormatTimestamp(time.Now())),
		ui.InfoStyle.Render(fmt.Sprintf("│ Message #%d │ Fichiers : %s", s.msgCount, s.sessionDir)),
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
	command := strings.ToLower(parts[0])

	switch command {
	case "/quit", "/exit", "/q":
		return true

	case "/help", "/h":
		printChatHelp()

	case "/clear", "/cls":
		system := s.history[0]
		s.history = []nimapi.Message{system}
		s.msgCount = 0
		s.titleSet = false
		ui.PrintSuccess("Historique effacé. Nouvelle conversation démarrée.")

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
		ui.PrintSuccess("Mode streaming " + status)

	case "/save":
		if err := s.saveHistory(); err != nil {
			ui.PrintError("Impossible de sauvegarder : " + err.Error())
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

	default:
		ui.PrintWarning("Commande inconnue : " + cmd + ". Tapez /help pour la liste.")
	}

	return false
}

func (s *ChatSession) handleSkillCommand(parts []string) {
	if len(parts) < 2 {
		if len(s.activeSkills) == 0 {
			fmt.Println(ui.InfoStyle.Render("Aucun skill actif."))
		} else {
			fmt.Println(ui.InfoStyle.Render("Skills actifs : ") + strings.Join(s.activeSkills, ", "))
		}
		return
	}

	sub := strings.ToLower(parts[1])
	switch sub {
	case "add":
		if len(parts) < 3 {
			ui.PrintError("Usage : /skill add <nom>")
			return
		}
		name := parts[2]
		sk, err := skills.Load(name)
		if err != nil {
			ui.PrintError(err.Error())
			return
		}
		for _, existing := range s.activeSkills {
			if existing == name {
				ui.PrintWarning("Skill déjà actif : " + name)
				return
			}
		}
		s.activeSkills = append(s.activeSkills, name)
		s.history[0] = nimapi.Message{
			Role:    "system",
			Content: buildSystemPrompt(s.activeSkills),
		}
		ui.PrintSuccess("Skill activé : " + sk.Name + " — " + sk.Description)

	case "remove", "rm":
		if len(parts) < 3 {
			ui.PrintError("Usage : /skill remove <nom>")
			return
		}
		name := parts[2]
		newSkills := make([]string, 0, len(s.activeSkills))
		found := false
		for _, sk := range s.activeSkills {
			if sk == name {
				found = true
				continue
			}
			newSkills = append(newSkills, sk)
		}
		if !found {
			ui.PrintWarning("Skill non actif : " + name)
			return
		}
		s.activeSkills = newSkills
		s.history[0] = nimapi.Message{
			Role:    "system",
			Content: buildSystemPrompt(s.activeSkills),
		}
		ui.PrintSuccess("Skill désactivé : " + name)

	case "list":
		allSkills, err := skills.List()
		if err != nil {
			ui.PrintError(err.Error())
			return
		}
		if len(allSkills) == 0 {
			dir, _ := skills.Dir()
			fmt.Println(ui.InfoStyle.Render("Aucun skill trouvé dans " + dir))
			return
		}
		fmt.Println()
		fmt.Println(ui.SectionTitleStyle.Render("  Skills disponibles"))
		fmt.Println(ui.Divider(50))
		for _, sk := range allSkills {
			active := "  "
			for _, a := range s.activeSkills {
				if a == sk.Name {
					active = ui.SuccessStyle.Render("✓ ")
					break
				}
			}
			fmt.Printf("%s%s  %s\n",
				active,
				ui.CommandStyle.Render(fmt.Sprintf("%-20s", sk.Name)),
				ui.InfoStyle.Render(sk.Description),
			)
		}
		fmt.Println()

	default:
		ui.PrintWarning("Sous-commande inconnue. Usage : /skill [add|remove|list]")
	}
}

func printCreatedFiles(files []filewriter.FileResult) {
	if len(files) == 0 {
		return
	}
	fmt.Println()
	for _, f := range files {
		if f.Err != nil {
			ui.PrintError(fmt.Sprintf("Fichier '%s' — %s", f.Name, f.Err.Error()))
		} else {
			ui.PrintSuccess(fmt.Sprintf("Fichier créé : %s", f.AbsPath))
		}
	}
}

func printChatWelcome(activeSkills []string) {
	line := ui.InfoStyle.Render("Session démarrée • ") +
		ui.CommandStyle.Render("/help") +
		ui.InfoStyle.Render(" pour les commandes • ") +
		ui.CommandStyle.Render("/quit") +
		ui.InfoStyle.Render(" pour quitter")

	if len(activeSkills) > 0 {
		line += "\n" + ui.InfoStyle.Render("Skills actifs : ") + ui.SuccessStyle.Render(strings.Join(activeSkills, ", "))
	}
	fmt.Println(line)
}

func printChatHelp() {
	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  Commandes disponibles dans le chat"))
	fmt.Println()

	cmds := []struct{ cmd, desc string }{
		{"/quit ou /q", "Quitter la session"},
		{"/clear", "Effacer l'historique de conversation"},
		{"/history", "Afficher l'historique de la session"},
		{"/model [id]", "Voir ou changer le modèle actif"},
		{"/stream", "Basculer le mode streaming on/off"},
		{"/system [msg]", "Voir ou modifier le message système"},
		{"/save", "Sauvegarder l'historique dans un fichier"},
		{"/title", "Afficher le titre et le dossier de la session"},
		{"/skill list", "Lister les skills disponibles"},
		{"/skill add <nom>", "Activer un skill"},
		{"/skill remove <nom>", "Désactiver un skill"},
		{"/help", "Afficher cette aide"},
	}

	for _, c := range cmds {
		fmt.Printf("%s %s\n",
			ui.CommandStyle.Render(fmt.Sprintf("  %-26s", c.cmd)),
			ui.InfoStyle.Render(c.desc),
		)
	}
	fmt.Println()
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

	filename := fmt.Sprintf("historique_%s.txt", time.Now().Format("20060102_150405"))
	path := s.sessionDir + "/" + filename

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Session NVIDIA NIM — %s\n", s.title)
	fmt.Fprintf(f, "Démarré : %s\n", s.started.Format("02/01/2006 15:04:05"))
	fmt.Fprintf(f, "Modèle : %s\n", s.model)
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
