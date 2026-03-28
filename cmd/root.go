package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/nvidia-nim-cli/internal/config"
	"github.com/user/nvidia-nim-cli/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "nim",
	Short: "Interface CLI pour NVIDIA AI Foundation Models",
	Long:  ui.Banner() + "\n" + ui.InfoStyle.Render("Utilisez 'nim --help' pour voir toutes les commandes disponibles."),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Banner())
		printQuickHelp()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if err := config.Init(); err != nil {
		ui.PrintWarning("Configuration : " + err.Error())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}

func printQuickHelp() {
	fmt.Println(ui.Divider(65))
	fmt.Println(ui.SectionTitleStyle.Render("  Commandes disponibles"))
	fmt.Println()

	commands := []struct{ cmd, desc string }{
		{"nim auth", "Configurer votre clé API NVIDIA"},
		{"nim chat", "Démarrer une session de chat interactive"},
		{"nim ask <question>", "Poser une question rapide sans mode interactif"},
		{"nim models", "Lister les modèles disponibles"},
		{"nim config", "Afficher et modifier la configuration"},
	}

	for _, c := range commands {
		cmdStr := ui.CommandStyle.Render(fmt.Sprintf("  %-32s", c.cmd))
		descStr := ui.InfoStyle.Render(c.desc)
		fmt.Printf("%s %s\n", cmdStr, descStr)
	}

	fmt.Println()
	fmt.Println(ui.Divider(65))
	fmt.Println(ui.InfoStyle.Render("  Commencez par : nim auth"))
	fmt.Println()
}
