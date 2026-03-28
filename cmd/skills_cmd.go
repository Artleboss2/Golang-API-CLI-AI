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
	skillNewDesc     string
	skillNewPrompt   string
	skillNewCategory string
	skillNewAuthor   string
)

func init() {
	skillsNewCmd.Flags().StringVarP(&skillNewDesc, "desc", "d", "", "Description courte du skill")
	skillsNewCmd.Flags().StringVarP(&skillNewPrompt, "prompt", "p", "", "Instructions système du skill")
	skillsNewCmd.Flags().StringVarP(&skillNewCategory, "category", "c", "", "Catégorie du skill")
	skillsNewCmd.Flags().StringVarP(&skillNewAuthor, "author", "a", "", "Auteur du skill")

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

	base, _ := skills.BaseDir()
	fmt.Println()
	fmt.Println(ui.SectionTitleStyle.Render("  ◆ Skills NVIDIA NIM CLI"))
	fmt.Println(ui.InfoStyle.Render("  Dossier : " + base))
	fmt.Println(ui.Divider(60))
	fmt.Println()

	if len(allSkills) == 0 {
		fmt.Println(ui.InfoStyle.Render("  Aucun skill trouvé."))
		fmt.Println()
		fmt.Println(ui.InfoStyle.Render("  Structure d'un skill :"))
		fmt.Println(ui.CommandStyle.Render("    " + base + "\\mon-skill\\"))
		fmt.Println(ui.InfoStyle.Render("      skill.yaml   ← métadonnées + prompt"))
		fmt.Println(ui.InfoStyle.Render("      README.md    ← contexte détaillé"))
		fmt.Println(ui.InfoStyle.Render("      assets\\      ← fichiers de référence"))
		fmt.Println()
		fmt.Println(ui.InfoStyle.Render("  Créer un skill :"))
		fmt.Println(ui.CommandStyle.Render(`    nim skills new mon-skill --desc "Expert Python" --prompt "Tu es un expert Python..."`))
		fmt.Println()
		return nil
	}

	for _, sk := range allSkills {
		fmt.Printf("  %s", ui.CommandStyle.Render(sk.Meta.Name))
		if sk.Meta.Version != "" {
			fmt.Printf("  %s", ui.TimestampStyle.Render("v"+sk.Meta.Version))
		}
		if sk.Meta.Category != "" {
			fmt.Printf("  %s", ui.ModelStyle.Render("["+sk.Meta.Category+"]"))
		}
		fmt.Println()

		if sk.Meta.Description != "" {
			fmt.Printf("  %s\n", ui.InfoStyle.Render(sk.Meta.Description))
		}
		if sk.Meta.Author != "" {
			fmt.Printf("  %s\n", ui.InfoStyle.Render("Auteur : "+sk.Meta.Author))
		}
		if len(sk.Meta.Tags) > 0 {
			fmt.Printf("  %s\n", ui.InfoStyle.Render("Tags : "+strings.Join(sk.Meta.Tags, ", ")))
		}

		hasReadme := sk.Readme != ""
		fmt.Printf("  %s  %s\n",
			ui.InfoStyle.Render("README.md :"),
			boolLabel(hasReadme),
		)
		fmt.Println()
	}

	fmt.Println(ui.Divider(60))
	fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  %d skill(s) • Activer avec : ", len(allSkills))) + ui.CommandStyle.Render("nim chat"))
	fmt.Println()
	return nil
}

func runSkillsNew(cmd *cobra.Command, args []string) error {
	name := args[0]

	if skillNewPrompt == "" {
		ui.PrintError("Le prompt est requis. Utilisez --prompt \"Tu es un expert en...\"")
		return nil
	}

	meta := skills.Meta{
		Name:        name,
		Description: skillNewDesc,
		Category:    skillNewCategory,
		Author:      skillNewAuthor,
		Version:     "1.0",
		Prompt:      skillNewPrompt,
	}

	if err := skills.Create(name, meta); err != nil {
		ui.PrintError("Impossible de créer le skill : " + err.Error())
		return nil
	}

	base, _ := skills.BaseDir()
	ui.PrintSuccess(fmt.Sprintf("Skill '%s' créé dans %s\\%s\\", name, base, name))
	fmt.Println(ui.InfoStyle.Render("  Fichiers créés :"))
	fmt.Println(ui.CommandStyle.Render(fmt.Sprintf("    %s\\%s\\skill.yaml", base, name)))
	fmt.Println(ui.CommandStyle.Render(fmt.Sprintf("    %s\\%s\\README.md", base, name)))
	fmt.Println(ui.CommandStyle.Render(fmt.Sprintf("    %s\\%s\\assets\\", base, name)))
	fmt.Println()
	fmt.Println(ui.InfoStyle.Render("  Activer au démarrage : ") + ui.CommandStyle.Render("nim chat"))
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
	fmt.Println(ui.SectionTitleStyle.Render("  Skill : " + sk.Meta.Name))
	if sk.Meta.Description != "" {
		fmt.Println(ui.InfoStyle.Render("  " + sk.Meta.Description))
	}
	fmt.Println(ui.Divider(60))

	if sk.Meta.Category != "" {
		fmt.Println(ui.InfoStyle.Render("  Catégorie : ") + sk.Meta.Category)
	}
	if sk.Meta.Author != "" {
		fmt.Println(ui.InfoStyle.Render("  Auteur    : ") + sk.Meta.Author)
	}
	if sk.Meta.Version != "" {
		fmt.Println(ui.InfoStyle.Render("  Version   : ") + sk.Meta.Version)
	}
	if len(sk.Meta.Tags) > 0 {
		fmt.Println(ui.InfoStyle.Render("  Tags      : ") + strings.Join(sk.Meta.Tags, ", "))
	}

	if sk.Meta.Prompt != "" {
		fmt.Println()
		fmt.Println(ui.CommandStyle.Render("  Prompt système :"))
		fmt.Println(sk.Meta.Prompt)
	}

	if sk.Readme != "" {
		fmt.Println()
		fmt.Println(ui.CommandStyle.Render("  README :"))
		lines := strings.Split(sk.Readme, "\n")
		limit := 20
		if len(lines) < limit {
			limit = len(lines)
		}
		for _, l := range lines[:limit] {
			fmt.Println("  " + l)
		}
		if len(lines) > 20 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("  ... (%d lignes supplémentaires)", len(lines)-20)))
		}
	}

	fmt.Println()
	return nil
}

func runSkillsDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Print(ui.WarningStyle.Render(fmt.Sprintf("Supprimer le skill '%s' et tous ses fichiers ? (oui/non) : ", name)))

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

func boolLabel(v bool) string {
	if v {
		return ui.SuccessStyle.Render("✓ présent")
	}
	return ui.InfoStyle.Render("absent")
}
