package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/nvidia-nim-cli/internal/skills"
	"github.com/user/nvidia-nim-cli/internal/ui"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Gérer les skills de l'assistant",
	Long:  `Liste, crée et supprime les skills qui personnalisent le comportement de l'IA.`,
	RunE:  runSkillsList,
}

var skillsNewCmd = &cobra.Command{
	Use:   "new <nom>",
	Short: "Créer un nouveau skill",
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsNew,
}

var skillsShowCmd = &cobra.Command{
	Use:   "show <nom>",
	Short: "Afficher le contenu d'un skill",
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsShow,
}

var skillsDeleteCmd = &cobra.Command{
	Use:   "delete <nom>",
	Short: "Supprimer un skill",
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsDelete,
}

var (
	skillNewDesc   string
	skillNewPrompt string
)

func init() {
	skillsNewCmd.Flags().StringVarP(&skillNewDesc, "desc", "d", "", "Description courte du skill")
	skillsNewCmd.Flags().StringVarP(&skillNewPrompt, "prompt", "p", "", "Contenu du prompt système")

	skillsCmd.AddCommand(skillsNewCmd)
	skillsCmd.AddCommand(skillsShowCmd)
	skillsCmd.AddCommand(skillsDeleteCmd)
}

func runSkillsList(cmd *cobra.Command, args []string) error {
	allSkills, err := skills.List()
	if err != nil {
		ui.PrintError(err.Error())
		return nil
	}

	dir, _ := skills.Dir()
	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  ◆ Skills NVIDIA NIM CLI"))
	fmt.Println(ui.InfoStyle.Render("  Dossier : " + dir))
	fmt.Println(ui.Divider(60))
	fmt.Println()

	if len(allSkills) == 0 {
		fmt.Println(ui.InfoStyle.Render("  Aucun skill trouvé."))
		fmt.Println()
		fmt.Println(ui.InfoStyle.Render("  Créer un skill :"))
		fmt.Println(ui.CommandStyle.Render(`    nim skills new mon-skill --desc "Mon expert" --prompt "Tu es un expert en..."`))
		fmt.Println()
		return nil
	}

	for _, sk := range allSkills {
		fmt.Printf("  %s\n", ui.CommandStyle.Render(sk.Name))
		fmt.Printf("  %s\n", ui.InfoStyle.Render(sk.Description))
		preview := sk.Prompt
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("  %s\n\n", ui.TimestampStyle.Render(preview))
	}

	fmt.Println(ui.Divider(60))
	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  %d skill(s) disponible(s)", len(allSkills))))
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("  Activer dans une session : ") + ui.CommandStyle.Render("nim chat -k <nom>"))
	fmt.Println()
	return nil
}

func runSkillsNew(cmd *cobra.Command, args []string) error {
	name := args[0]

	desc := skillNewDesc
	if desc == "" {
		desc = name
	}

	prompt := skillNewPrompt
	if prompt == "" {
		ui.PrintError("Le prompt est requis. Utilisez --prompt \"Tu es un expert en...\"")
		return nil
	}

	content := fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n", desc, prompt)

	if err := skills.Save(name, content); err != nil {
		ui.PrintError("Impossible de créer le skill : " + err.Error())
		return nil
	}

	dir, _ := skills.Dir()
	ui.PrintSuccess(fmt.Sprintf("Skill '%s' créé dans %s/%s.md", name, dir, name))
	fmt.Println(ui.InfoStyle.Render("  Activer avec : ") + ui.CommandStyle.Render("nim chat -k "+name))
	return nil
}

func runSkillsShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	sk, err := skills.Load(name)
	if err != nil {
		ui.PrintError(err.Error())
		return nil
	}

	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  Skill : " + sk.Name))
	fmt.Println(ui.InfoStyle.Render("  " + sk.Description))
	fmt.Println(ui.Divider(60))
	fmt.Println()
	fmt.Println(sk.Prompt)
	fmt.Println()
	return nil
}

func runSkillsDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Print(ui.WarningStyle.Render(fmt.Sprintf("Supprimer le skill '%s' ? (oui/non) : ", name)))

	var confirm string
	fmt.Fscan(os.Stdin, &confirm)

	if !strings.EqualFold(confirm, "oui") && !strings.EqualFold(confirm, "o") {
		fmt.Println(ui.InfoStyle.Render("Annulé."))
		return nil
	}

	if err := skills.Delete(name); err != nil {
		ui.PrintError(err.Error())
		return nil
	}

	ui.PrintSuccess(fmt.Sprintf("Skill '%s' supprimé.", name))
	return nil
}
