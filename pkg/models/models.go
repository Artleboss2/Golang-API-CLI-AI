package models

type Model struct {
	ID          string
	DisplayName string
	Provider    string
	ContextLen  int
	Description string
	Speed       string
	Specialty   string
}

var Available = []Model{
	{
		ID:          "meta/llama3-70b-instruct",
		DisplayName: "Llama 3 70B Instruct",
		Provider:    "Meta",
		ContextLen:  8192,
		Description: "Modèle phare de Meta, excellent pour les tâches complexes",
		Speed:       "⚡⚡",
		Specialty:   "Raisonnement, Code, Analyse",
	},
	{
		ID:          "meta/llama3-8b-instruct",
		DisplayName: "Llama 3 8B Instruct",
		Provider:    "Meta",
		ContextLen:  8192,
		Description: "Version légère de Llama 3, plus rapide et efficace",
		Speed:       "⚡⚡⚡",
		Specialty:   "Conversation rapide, Résumé",
	},
	{
		ID:          "meta/llama-3.1-405b-instruct",
		DisplayName: "Llama 3.1 405B Instruct",
		Provider:    "Meta",
		ContextLen:  128000,
		Description: "Le plus puissant de la famille LLaMA — contexte 128k",
		Speed:       "⚡",
		Specialty:   "Tâches complexes, Recherche",
	},
	{
		ID:          "meta/llama-3.1-70b-instruct",
		DisplayName: "Llama 3.1 70B Instruct",
		Provider:    "Meta",
		ContextLen:  128000,
		Description: "LLaMA 3.1 avec fenêtre de contexte étendue",
		Speed:       "⚡⚡",
		Specialty:   "Documents longs, Code",
	},
	{
		ID:          "meta/llama-3.1-8b-instruct",
		DisplayName: "Llama 3.1 8B Instruct",
		Provider:    "Meta",
		ContextLen:  128000,
		Description: "Léger et rapide avec support 128k tokens",
		Speed:       "⚡⚡⚡",
		Specialty:   "Usage général rapide",
	},
	{
		ID:          "mistralai/mixtral-8x7b-instruct-v0.1",
		DisplayName: "Mixtral 8x7B Instruct",
		Provider:    "Mistral AI",
		ContextLen:  32768,
		Description: "Architecture MoE efficace, excellent rapport qualité/vitesse",
		Speed:       "⚡⚡⚡",
		Specialty:   "Multilingue, Code, Raisonnement",
	},
	{
		ID:          "mistralai/mistral-7b-instruct-v0.3",
		DisplayName: "Mistral 7B Instruct v0.3",
		Provider:    "Mistral AI",
		ContextLen:  32768,
		Description: "Petit mais puissant, idéal pour le déploiement rapide",
		Speed:       "⚡⚡⚡",
		Specialty:   "Chat léger, Instructions",
	},
	{
		ID:          "mistralai/mistral-large",
		DisplayName: "Mistral Large",
		Provider:    "Mistral AI",
		ContextLen:  128000,
		Description: "Le plus grand modèle Mistral, rival des SOTA",
		Speed:       "⚡⚡",
		Specialty:   "Raisonnement avancé, Code complexe",
	},
	{
		ID:          "mistralai/mixtral-8x22b-instruct-v0.1",
		DisplayName: "Mixtral 8x22B Instruct",
		Provider:    "Mistral AI",
		ContextLen:  65536,
		Description: "MoE géant pour les tâches professionnelles exigeantes",
		Speed:       "⚡",
		Specialty:   "Analyse approfondie, Recherche",
	},
	{
		ID:          "microsoft/phi-3-mini-128k-instruct",
		DisplayName: "Phi-3 Mini 128K",
		Provider:    "Microsoft",
		ContextLen:  128000,
		Description: "Modèle compact ultra-performant de Microsoft",
		Speed:       "⚡⚡⚡",
		Specialty:   "Efficacité, Déploiement edge",
	},
	{
		ID:          "microsoft/phi-3-medium-128k-instruct",
		DisplayName: "Phi-3 Medium 128K",
		Provider:    "Microsoft",
		ContextLen:  128000,
		Description: "Équilibre parfait taille/performance",
		Speed:       "⚡⚡⚡",
		Specialty:   "Raisonnement, Instructions longues",
	},
	{
		ID:          "google/gemma-7b",
		DisplayName: "Gemma 7B",
		Provider:    "Google",
		ContextLen:  8192,
		Description: "Modèle ouvert léger de Google DeepMind",
		Speed:       "⚡⚡⚡",
		Specialty:   "Conversation, Résumé",
	},
	{
		ID:          "google/codegemma-7b",
		DisplayName: "CodeGemma 7B",
		Provider:    "Google",
		ContextLen:  8192,
		Description: "Spécialisé pour la génération et analyse de code",
		Speed:       "⚡⚡⚡",
		Specialty:   "Code, Débogage",
	},
	{
		ID:          "deepseek-ai/deepseek-coder-6.7b-instruct",
		DisplayName: "DeepSeek Coder 6.7B",
		Provider:    "DeepSeek",
		ContextLen:  16384,
		Description: "Expert en programmation multi-langages",
		Speed:       "⚡⚡⚡",
		Specialty:   "Code, Algorithmes",
	},
}

func FindByID(id string) (Model, bool) {
	for _, m := range Available {
		if m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}

func GetDisplayName(id string) string {
	if m, found := FindByID(id); found {
		return m.DisplayName
	}
	return id
}

func GetIDs() []string {
	ids := make([]string, len(Available))
	for i, m := range Available {
		ids[i] = m.ID
	}
	return ids
}
