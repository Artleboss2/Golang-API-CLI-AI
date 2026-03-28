package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/ui"
	"github.com/user/nvidia-nim-cli/pkg/models"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Afficher et modifier la configuration du CLI",
	Long:  `Affiche la configuration actuelle et permet de modifier les paramètres.`,
	RunE:  runConfig,
}

var configSetCmd = &cobra.Command{
	Use:   "set <clé> <valeur>",
	Short: "Modifier un paramètre de configuration",
	Long: `Modifie un paramètre de configuration.

Clés disponibles :
  model         Modèle NIM par défaut
  temperature   Température (0.0–1.0)
  max-tokens    Nombre maximum de tokens
  stream        Mode streaming (true/false)`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  ◆ Configuration NVIDIA NIM CLI"))
	fmt.Println(ui.Divider(60))
	fmt.Println()

	cfg, err := config.GetAll()
	if err != nil {
		ui.PrintError("Impossible de lire la configuration : " + err.Error())
		return nil
	}

	printConfigRow("Fichier config", config.GetConfigPath())
	fmt.Println()

	if cfg.APIKey == "" {
		printConfigRow("Clé API", ui.WarningStyle.Render("⚠ Non configurée — Utilisez 'nim auth'"))
	} else {
		printConfigRow("Clé API", maskAPIKey(cfg.APIKey)+ui.SuccessStyle.Render(" ✓"))
	}

	printConfigRow("Modèle par défaut", ui.ModelStyle.Render(models.GetDisplayName(cfg.DefaultModel)))
	printConfigRow("  └─ ID", ui.InfoStyle.Render(cfg.DefaultModel))

	fmt.Println()
	printConfigRow("Température", fmt.Sprintf("%.1f", cfg.Temperature))
	printConfigRow("Tokens maximum", fmt.Sprintf("%d", cfg.MaxTokens))

	streamStatus := ui.SuccessStyle.Render("✓ Activé")
	if !cfg.Stream {
		streamStatus = ui.InfoStyle.Render("✗ Désactivé")
	}
	printConfigRow("Streaming", streamStatus)

	fmt.Println()
	fmt.Println(ui.Divider(60))
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("  Pour modifier un paramètre :"))
	fmt.Println(ui.CommandStyle.Render("    nim config set temperature 0.9"))
	fmt.Println(ui.CommandStyle.Render("    nim config set model meta/llama3-8b-instruct"))
	fmt.Println(ui.CommandStyle.Render("    nim config set max-tokens 2048"))
	fmt.Println(ui.CommandStyle.Render("    nim config set stream false"))
	fmt.Println()

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	switch key {
	case "model":
		if err := config.SaveModel(value); err != nil {
			ui.PrintError("Impossible de sauvegarder : " + err.Error())
			return nil
		}
		ui.PrintSuccess("Modèle mis à jour : " + models.GetDisplayName(value))

	case "temperature", "temp":
		var temp float64
		if _, err := fmt.Sscanf(value, "%f", &temp); err != nil || temp < 0 || temp > 2 {
			ui.PrintError("Température invalide. Valeur entre 0.0 et 2.0 attendue.")
			return nil
		}
		cfg, _ := config.GetAll()
		cfg.Temperature = temp
		if err := config.SaveAll(cfg); err != nil {
			ui.PrintError("Impossible de sauvegarder : " + err.Error())
			return nil
		}
		ui.PrintSuccess(fmt.Sprintf("Température mise à jour : %.1f", temp))

	case "max-tokens", "max_tokens", "tokens":
		var tokens int
		if _, err := fmt.Sscanf(value, "%d", &tokens); err != nil || tokens <= 0 {
			ui.PrintError("Nombre de tokens invalide. Entier positif attendu.")
			return nil
		}
		cfg, _ := config.GetAll()
		cfg.MaxTokens = tokens
		if err := config.SaveAll(cfg); err != nil {
			ui.PrintError("Impossible de sauvegarder : " + err.Error())
			return nil
		}
		ui.PrintSuccess(fmt.Sprintf("Tokens maximum mis à jour : %d", tokens))

	case "stream", "streaming":
		streamEnabled := value == "true" || value == "1" || value == "on" || value == "oui"
		cfg, _ := config.GetAll()
		cfg.Stream = streamEnabled
		if err := config.SaveAll(cfg); err != nil {
			ui.PrintError("Impossible de sauvegarder : " + err.Error())
			return nil
		}
		status := "désactivé"
		if streamEnabled {
			status = "activé"
		}
		ui.PrintSuccess("Streaming " + status)

	default:
		ui.PrintError("Clé inconnue : " + key)
		fmt.Println(ui.InfoStyle.Render("Clés valides : model, temperature, max-tokens, stream"))
	}

	return nil
}

func printConfigRow(key, value string) {
	fmt.Printf("%s %s\n",
		ui.InfoStyle.Render(fmt.Sprintf("  %-22s", key)),
		value,
	)
}
