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
	chatModel  string
	chatSystem string
	chatStream bool
	chatTemp   float64
)

func init() {
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "", "Modèle NIM à utiliser")
	chatCmd.Flags().StringVarP(&chatSystem, "system", "s", "", "Message système personnalisé")
	chatCmd.Flags().BoolVar(&chatStream, "stream", true, "Activer le streaming des réponses")
	chatCmd.Flags().Float64VarP(&chatTemp, "temperature", "t", -1, "Température (0.0–1.0)")
}

type ChatSession struct {
	client   *nimapi.Client
	model    string
	history  []nimapi.Message
	stream   bool
	temp     float64
	msgCount int
	started  time.Time
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
		client:  nimapi.NewClient(config.GetAPIKey()),
		model:   model,
		stream:  chatStream,
		temp:    temperature,
		started: time.Now(),
	}

	systemMsg := chatSystem
	if systemMsg == "" {
		systemMsg = "Tu es un assistant IA intelligent et concis. Réponds en français sauf si l'utilisateur parle dans une autre langue. Sois précis, utile et professionnel."
	}

	session.history = []nimapi.Message{
		{Role: "system", Content: systemMsg},
	}

	return session.run()
}

func (s *ChatSession) run() error {
	fmt.Println()
	fmt.Println(ui.SmallBanner(models.GetDisplayName(s.model)))
	fmt.Println()
	printChatWelcome()

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

	if s.stream {
		tokenChan := make(chan string, 100)
		errChan := make(chan error, 1)

		go s.client.StreamComplete(req, tokenChan, errChan)

	loop:
		for {
			select {
			case token, ok := <-tokenChan:
				if !ok {
					break loop
				}
				fmt.Print(ui.AIMessageStyle.Render(token))
				fullResponse += token
			case err := <-errChan:
				if err != nil {
					fmt.Println()
					return fmt.Errorf("erreur streaming : %w", err)
				}
			}
		}
		fmt.Println()
	} else {
		spinner := ui.NewSpinner("Génération en cours...")
		spinner.Start()

		resp, err := s.client.Complete(req)
		if err != nil {
			spinner.StopWithError("")
			return err
		}
		spinner.Stop("")

		if len(resp.Choices) > 0 {
			fullResponse = resp.Choices[0].Message.Content
			fmt.Println(ui.AIMessageStyle.Render(fullResponse))
		}
	}

	if fullResponse != "" {
		s.history = append(s.history, nimapi.Message{
			Role:    "assistant",
			Content: fullResponse,
		})
	}

	fmt.Printf("\n%s %s\n",
		ui.TimestampStyle.Render(ui.FormatTimestamp(time.Now())),
		ui.InfoStyle.Render(fmt.Sprintf("│ Message #%d │ Tapez /help pour les commandes", s.msgCount)),
	)

	return nil
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
			ui.PrintSuccess("Historique sauvegardé.")
		}

	case "/system":
		if len(parts) > 1 {
			s.history[0] = nimapi.Message{Role: "system", Content: strings.Join(parts[1:], " ")}
			ui.PrintSuccess("Message système mis à jour.")
		} else {
			fmt.Println(ui.InfoStyle.Render("Système actuel : ") + s.history[0].Content)
		}

	default:
		ui.PrintWarning("Commande inconnue : " + cmd + ". Tapez /help pour la liste.")
	}

	return false
}

func printChatWelcome() {
	fmt.Println(
		ui.InfoStyle.Render("Session démarrée • Tapez ") +
			ui.CommandStyle.Render("/help") +
			ui.InfoStyle.Render(" pour les commandes • ") +
			ui.CommandStyle.Render("/quit") +
			ui.InfoStyle.Render(" pour quitter"),
	)
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
		{"/help", "Afficher cette aide"},
	}

	for _, c := range cmds {
		fmt.Printf("%s %s\n",
			ui.CommandStyle.Render(fmt.Sprintf("  %-20s", c.cmd)),
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
	filename := fmt.Sprintf("chat_%s.txt", time.Now().Format("20060102_150405"))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Session NVIDIA NIM — %s\n", s.started.Format("02/01/2006 15:04:05"))
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
	fmt.Println(ui.Divider(60))
	fmt.Println()
}
