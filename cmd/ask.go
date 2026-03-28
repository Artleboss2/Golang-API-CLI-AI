package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	nimapi "github.com/user/nvidia-nim-cli/internal/api"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/filewriter"
	"github.com/user/nvidia-nim-cli/internal/skills"
	"github.com/user/nvidia-nim-cli/internal/titler"
	"github.com/user/nvidia-nim-cli/internal/ui"
	"github.com/user/nvidia-nim-cli/pkg/models"
)

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Poser une question rapide au modèle NIM",
	Long: `Envoie une question unique au modèle et affiche la réponse.

Exemples :
  nim ask "Explique le fonctionnement d'un GPU"
  nim ask --model meta/llama3-8b-instruct "Traduis 'hello world' en français"
  nim ask --no-stream "Liste 5 langages de programmation"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAsk,
}

var (
	askModel    string
	askTemp     float64
	askMaxTok   int
	askSystem   string
	askNoStream bool
	askSkills   []string
)

func init() {
	askCmd.Flags().StringVarP(&askModel, "model", "m", "", "Modèle NIM à utiliser")
	askCmd.Flags().Float64VarP(&askTemp, "temperature", "t", -1, "Température (0.0–1.0)")
	askCmd.Flags().IntVar(&askMaxTok, "max-tokens", 0, "Nombre maximum de tokens")
	askCmd.Flags().StringVarP(&askSystem, "system", "s", "", "Message système personnalisé")
	askCmd.Flags().BoolVar(&askNoStream, "no-stream", false, "Désactiver le streaming")
	askCmd.Flags().StringSliceVarP(&askSkills, "skill", "k", nil, "Skills à activer")
}

func runAsk(cmd *cobra.Command, args []string) error {
	if !config.HasAPIKey() {
		ui.PrintError("Aucune clé API configurée.")
		fmt.Println(ui.InfoStyle.Render("Configurez votre clé avec : ") + ui.CommandStyle.Render("nim auth"))
		return nil
	}

	question := strings.Join(args, " ")

	model := askModel
	if model == "" {
		model = config.GetModel()
	}

	temperature := askTemp
	if temperature < 0 {
		temperature = config.GetTemperature()
	}

	maxTokens := askMaxTok
	if maxTokens <= 0 {
		maxTokens = config.GetMaxTokens()
	}

	useStream := config.IsStreamEnabled() && !askNoStream

	systemMsg := askSystem
	if systemMsg == "" {
		systemMsg = buildAskSystemPrompt(askSkills)
	}

	sessionTitle := titler.Fallback()
	sessionDir := filewriter.SessionDir(sessionTitle)

	client := nimapi.NewClient(config.GetAPIKey())

	req := &nimapi.ChatRequest{
		Model: model,
		Messages: []nimapi.Message{
			{Role: "system", Content: systemMsg},
			{Role: "user", Content: question},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      useStream,
	}

	fmt.Println()
	fmt.Println(ui.UserPrefixStyle.Render("Vous ▶ ") + ui.UserMessageStyle.Render(question))
	fmt.Println()
	fmt.Print(ui.AIPrefixStyle.Render("◆ " + models.GetDisplayName(model) + " ▶ "))
	fmt.Println()
	fmt.Println()

	if useStream {
		return sendStreamingAsk(client, req, sessionDir)
	}
	return sendStandardAsk(client, req, sessionDir)
}

func buildAskSystemPrompt(activeSkills []string) string {
	base := `Tu es un assistant IA utile et concis. Réponds en français sauf si la question est dans une autre langue. Sois précis et direct.

Tu peux créer des fichiers directement sur le PC de l'utilisateur avec ce format :
<file:nom_du_fichier.ext>
contenu du fichier
</file>`

	addendum := skills.BuildSystemAddendum(activeSkills)
	return base + addendum
}

func sendStreamingAsk(client *nimapi.Client, req *nimapi.ChatRequest, sessionDir string) error {
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go client.StreamComplete(req, tokenChan, errChan)

	var sb strings.Builder

	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				fmt.Println()
				fullResponse := sb.String()
				_, createdFiles := filewriter.ExtractAndWrite(fullResponse, sessionDir)
				printCreatedFiles(createdFiles)
				fmt.Println()
				return nil
			}
			fmt.Print(ui.AIMessageStyle.Render(token))
			sb.WriteString(token)
		case err := <-errChan:
			if err != nil {
				fmt.Println()
				return fmt.Errorf("erreur streaming : %w", err)
			}
		}
	}
}

func sendStandardAsk(client *nimapi.Client, req *nimapi.ChatRequest, sessionDir string) error {
	spinner := ui.NewSpinner("Génération de la réponse...")
	spinner.Start()

	resp, err := client.Complete(req)
	if err != nil {
		spinner.StopWithError("")
		return err
	}
	spinner.Stop("")

	if len(resp.Choices) == 0 {
		return nil
	}

	raw := resp.Choices[0].Message.Content
	cleanResponse, createdFiles := filewriter.ExtractAndWrite(raw, sessionDir)

	fmt.Println(ui.AIMessageStyle.Render(cleanResponse))
	fmt.Println()

	printCreatedFiles(createdFiles)

	if resp.Usage.TotalTokens > 0 {
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf(
			"Tokens utilisés : %d (prompt: %d, réponse: %d)",
			resp.Usage.TotalTokens,
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
		)))
	}

	return nil
}
