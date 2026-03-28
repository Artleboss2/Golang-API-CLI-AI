package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/user/nvidia-nim-cli/internal/api"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/ui"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configurer la clé API NVIDIA NIM",
	Long: `Configure et sauvegarde votre clé API NVIDIA NIM localement.

La clé sera stockée dans ~/.nvidia-nim/config.yaml
Obtenez votre clé gratuite sur : https://build.nvidia.com`,
	RunE: runAuth,
}

var (
	authAPIKey   string
	authValidate bool
	authShow     bool
)

func init() {
	authCmd.Flags().StringVarP(&authAPIKey, "key", "k", "", "Clé API à sauvegarder directement")
	authCmd.Flags().BoolVarP(&authValidate, "validate", "v", true, "Valider la clé contre l'API NVIDIA")
	authCmd.Flags().BoolVarP(&authShow, "show", "s", false, "Afficher la clé API actuelle")
}

func runAuth(cmd *cobra.Command, args []string) error {
	fmt.Println()

	if authShow {
		return showCurrentKey()
	}

	printAuthHeader()

	apiKey := authAPIKey
	if apiKey == "" {
		var err error
		apiKey, err = promptAPIKey()
		if err != nil {
			return fmt.Errorf("saisie annulée : %w", err)
		}
	}

	if strings.TrimSpace(apiKey) == "" {
		ui.PrintError("La clé API ne peut pas être vide.")
		return nil
	}

	if authValidate {
		if err := validateKey(apiKey); err != nil {
			return nil
		}
	}

	if err := config.SaveAPIKey(apiKey); err != nil {
		ui.PrintError("Impossible de sauvegarder la clé : " + err.Error())
		return nil
	}

	printAuthSuccess(apiKey)
	return nil
}

func printAuthHeader() {
	header := ui.BoxMessage(
		"◆ Authentification NVIDIA NIM",
		ui.InfoStyle.Render("Votre clé API est disponible sur :")+"\n"+
			ui.CommandStyle.Render("  → https://build.nvidia.com/explore/discover")+"\n\n"+
			ui.InfoStyle.Render("Elle sera sauvegardée dans : ")+
			ui.PromptStyle.Render(config.GetConfigPath()),
	)
	fmt.Println(header)
	fmt.Println()
}

func promptAPIKey() (string, error) {
	fmt.Print(ui.PromptStyle.Render("  Clé API NVIDIA ▶ "))

	if term.IsTerminal(int(syscall.Stdin)) {
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(password)), nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", scanner.Err()
}

func validateKey(apiKey string) error {
	spinner := ui.NewSpinner("Validation de la clé API en cours...")
	spinner.Start()

	client := api.NewClient(apiKey)
	err := client.ValidateAPIKey(config.GetModel())

	if err != nil {
		spinner.StopWithError("Clé API invalide : " + err.Error())
		fmt.Println()
		ui.PrintWarning("Vérifiez votre clé sur https://build.nvidia.com")
		fmt.Println()
		return err
	}

	spinner.Stop("Clé API validée avec succès !")
	return nil
}

func showCurrentKey() error {
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		ui.PrintWarning("Aucune clé API configurée. Utilisez 'nim auth' pour en ajouter une.")
		return nil
	}

	masked := maskAPIKey(apiKey)
	fmt.Println(ui.InfoStyle.Render("Clé API actuelle : ") + ui.SuccessStyle.Render(masked))
	fmt.Println(ui.InfoStyle.Render("Fichier config   : ") + config.GetConfigPath())
	return nil
}

func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + strings.Repeat("*", len(key)-12) + key[len(key)-4:]
}

func printAuthSuccess(apiKey string) {
	fmt.Println()
	fmt.Println(ui.SuccessStyle.Render("✓ Clé API sauvegardée avec succès !"))
	fmt.Println(ui.InfoStyle.Render("  Clé     : ") + maskAPIKey(apiKey))
	fmt.Println(ui.InfoStyle.Render("  Fichier : ") + config.GetConfigPath())
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("Prochaine étape : ") + ui.CommandStyle.Render("nim chat"))
	fmt.Println()
}
