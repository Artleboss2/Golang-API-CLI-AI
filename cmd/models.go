package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/ui"
	"github.com/user/nvidia-nim-cli/pkg/models"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Lister les modèles NVIDIA NIM disponibles",
	Long:  `Affiche le catalogue complet des modèles accessibles via l'API NVIDIA NIM.`,
	RunE:  runModels,
}

var (
	modelsProvider string
	modelsSet      string
)

func init() {
	modelsCmd.Flags().StringVarP(&modelsProvider, "provider", "p", "", "Filtrer par fournisseur (ex: Meta, Mistral)")
	modelsCmd.Flags().StringVarP(&modelsSet, "set", "s", "", "Définir le modèle par défaut")
}

func runModels(cmd *cobra.Command, args []string) error {
	if modelsSet != "" {
		return setDefaultModel(modelsSet)
	}

	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  ◆ Catalogue des modèles NVIDIA NIM"))
	fmt.Println()

	currentModel := config.GetModel()

	filteredModels := models.Available
	if modelsProvider != "" {
		filteredModels = filterByProvider(models.Available, modelsProvider)
		if len(filteredModels) == 0 {
			ui.PrintWarning("Aucun modèle trouvé pour le fournisseur : " + modelsProvider)
			return nil
		}
	}

	groups := groupByProvider(filteredModels)
	providers := getOrderedProviders(filteredModels)

	for _, provider := range providers {
		fmt.Println(ui.ModelStyle.Render("  ▸ " + provider))
		fmt.Println(ui.Divider(70))

		for _, m := range groups[provider] {
			printModelRow(m, currentModel)
		}

		fmt.Println()
	}

	fmt.Println(ui.Divider(70))
	fmt.Println(
		ui.InfoStyle.Render(fmt.Sprintf("  %d modèles disponibles │ Modèle actuel : ", len(filteredModels)))+
			ui.ModelStyle.Render(models.GetDisplayName(currentModel)),
	)
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("  Pour changer de modèle : ") +
		ui.CommandStyle.Render("nim models --set <model-id>"))
	fmt.Println()

	return nil
}

func printModelRow(m models.Model, currentModel string) {
	activeMarker := "  "
	if m.ID == currentModel {
		activeMarker = ui.SuccessStyle.Render("✓ ")
	}

	nameStyle := ui.AppNameStyle
	if m.ID == currentModel {
		nameStyle = ui.SuccessStyle
	}

	fmt.Printf("  %s%s %s %s\n",
		activeMarker,
		nameStyle.Render(fmt.Sprintf("%-30s", m.DisplayName)),
		ui.InfoStyle.Render(fmt.Sprintf("%-10s", formatContext(m.ContextLen))),
		m.Speed,
	)
	fmt.Printf("      %s\n", ui.InfoStyle.Render(m.Description))
	fmt.Printf("      %s\n", ui.TimestampStyle.Render("ID : "+m.ID))
	fmt.Println()
}

func setDefaultModel(modelID string) error {
	if _, found := models.FindByID(modelID); !found {
		ui.PrintWarning("Modèle '" + modelID + "' non trouvé dans le catalogue local.")
		ui.PrintInfo("Vérifiez l'ID exact sur https://build.nvidia.com/explore/discover")
	}

	if err := config.SaveModel(modelID); err != nil {
		ui.PrintError("Impossible de sauvegarder : " + err.Error())
		return nil
	}

	ui.PrintSuccess("Modèle par défaut mis à jour : " + models.GetDisplayName(modelID))
	return nil
}

func filterByProvider(modelList []models.Model, provider string) []models.Model {
	var result []models.Model
	providerLower := strings.ToLower(provider)
	for _, m := range modelList {
		if strings.Contains(strings.ToLower(m.Provider), providerLower) {
			result = append(result, m)
		}
	}
	return result
}

func groupByProvider(modelList []models.Model) map[string][]models.Model {
	groups := make(map[string][]models.Model)
	for _, m := range modelList {
		groups[m.Provider] = append(groups[m.Provider], m)
	}
	return groups
}

func getOrderedProviders(modelList []models.Model) []string {
	seen := make(map[string]bool)
	var result []string
	for _, m := range modelList {
		if !seen[m.Provider] {
			seen[m.Provider] = true
			result = append(result, m.Provider)
		}
	}
	return result
}

func formatContext(tokens int) string {
	if tokens >= 1000 {
		return fmt.Sprintf("%dk ctx", tokens/1000)
	}
	return fmt.Sprintf("%d ctx", tokens)
}
