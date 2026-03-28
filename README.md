# NVIDIA NIM CLI

> Interface en ligne de commande pour les modèles NVIDIA AI Foundation Models (NIM) — rapide, élégante et puissante.

```
 ███╗   ██╗██╗   ██╗██╗██████╗ ██╗ █████╗     ███╗   ██╗██╗███╗   ███╗
 ████╗  ██║██║   ██║██║██╔══██╗██║██╔══██╗    ████╗  ██║██║████╗ ████║
 ██╔██╗ ██║██║   ██║██║██║  ██║██║███████║    ██╔██╗ ██║██║██╔████╔██║
 ██║╚██╗██║╚██╗ ██╔╝██║██║  ██║██║██╔══██║    ██║╚██╗██║██║██║╚██╔╝██║
 ██║ ╚████║ ╚████╔╝ ██║██████╔╝██║██║  ██║    ██║ ╚████║██║██║ ╚═╝ ██║
 ╚═╝  ╚═══╝  ╚═══╝  ╚═╝╚═════╝ ╚═╝╚═╝  ╚═╝    ╚═╝  ╚═══╝╚═╝╚═╝     ╚═╝
```

\---

## &#x20;Fonctionnalités

|Fonctionnalité|Description|
|-|-|
|**`nim auth`**|Sauvegarde sécurisée de votre clé API NVIDIA|
|**`nim chat`**|Session interactive avec historique de conversation|
|**`nim ask`**|Question rapide one-shot (idéal pour scripts)|
|**`nim models`**|Catalogue complet des modèles disponibles|
|**`nim config`**|Gestion fine des paramètres (température, tokens...)|
|**Streaming SSE**|Réponses en temps réel (effet machine à écrire)|
|**14+ modèles**|Llama 3, Mixtral, Mistral, Phi-3, Gemma, DeepSeek...|

\---

## Installation sur Windows

### Prérequis

1. **Installer Go 1.22+**

   * Téléchargez depuis [https://go.dev/dl/](https://go.dev/dl/)
   * Choisissez `go1.22.x.windows-amd64.msi`
   * Suivez l'installateur (redémarrez le terminal après)
   * Vérifiez : `go version`
2. **Obtenir une clé API NVIDIA (gratuite)**

   * Créez un compte sur [https://build.nvidia.com](https://build.nvidia.com)
   * Allez dans **"Get API Key"**
   * Copiez votre clé (commence par `nvapi-...`)

### Compilation et Installation

```batch
:: 1. Cloner ou extraire le projet
cd C:\\Users\\VotreNom\\nvidia-nim-cli

:: 2. Télécharger les dépendances et compiler
build.bat

:: 3a. Installation globale (accès depuis partout)
::     Copiez nim.exe dans un dossier déjà dans votre PATH
move nim.exe C:\\Windows\\System32\\

:: 3b. OU ajoutez le dossier au PATH Windows :
::     Panneau de configuration > Variables d'environnement
::     > PATH > Nouveau > C:\\Users\\VotreNom\\nvidia-nim-cli
```

### Vérification

```batch
nim --version
:: Doit afficher : nvidia-nim-cli v1.0.0
```

\---

## Démarrage rapide

```batch
:: Étape 1 : Configurer votre clé API
nim auth
:: (Entrez votre clé nvapi-... quand demandé)

:: Étape 2 : Démarrer un chat
nim chat

:: Étape 3 : Ou poser une question rapide
nim ask "Explique-moi le machine learning en 3 phrases"
```

\---

## Guide des commandes

### `nim auth` — Authentification

```batch
nim auth                     # Saisie interactive de la clé
nim auth --key nvapi-xxxxx   # Passer la clé directement
nim auth --show              # Afficher la clé actuelle (masquée)
nim auth --no-validate       # Sauvegarder sans vérifier
```

La clé est sauvegardée dans `%USERPROFILE%\\.nvidia-nim\\config.yaml`

\---

### `nim chat` — Mode interactif

```batch
nim chat                                    # Démarrer un chat
nim chat --model meta/llama3-8b-instruct    # Choisir un modèle
nim chat --temperature 0.9                  # Plus créatif
nim chat --no-stream                        # Sans streaming
nim chat --system "Tu es un expert Python"  # Personnaliser l'IA
```

**Commandes dans le chat :**

|Commande|Action|
|-|-|
|`/help`|Afficher l'aide|
|`/quit` ou `/q`|Quitter la session|
|`/clear`|Effacer l'historique|
|`/history`|Voir la conversation|
|`/model meta/llama3-8b-instruct`|Changer de modèle|
|`/stream`|Activer/désactiver le streaming|
|`/save`|Sauvegarder la conversation|
|`/system \[message]`|Modifier l'instruction système|

\---

### `nim ask` — Question rapide

```batch
nim ask "Quel est le meilleur algorithme de tri ?"
nim ask --model mistralai/mistral-large "Traduis en anglais : Bonjour"
nim ask --no-stream --max-tokens 500 "Liste les frameworks Go populaires"
nim ask --temperature 1.0 "Écris un haïku sur la programmation"
```

\---

### `nim models` — Catalogue des modèles

```batch
nim models                          # Lister tous les modèles
nim models --provider Meta          # Filtrer par Meta
nim models --provider Mistral       # Filtrer par Mistral
nim models --set meta/llama3-70b-instruct  # Définir le modèle par défaut
```

**Modèles disponibles :**

|Modèle|Provider|Contexte|Usage|
|-|-|-|-|
|`meta/llama3-70b-instruct`|Meta|8k|Raisonnement, Code|
|`meta/llama3-8b-instruct`|Meta|8k|Chat rapide|
|`meta/llama-3.1-405b-instruct`|Meta|128k|Tâches complexes|
|`meta/llama-3.1-70b-instruct`|Meta|128k|Documents longs|
|`mistralai/mixtral-8x7b-instruct-v0.1`|Mistral AI|32k|Multilingue|
|`mistralai/mistral-large`|Mistral AI|128k|Analyse avancée|
|`microsoft/phi-3-mini-128k-instruct`|Microsoft|128k|Edge, efficacité|
|`google/gemma-7b`|Google|8k|Conversation|
|`google/codegemma-7b`|Google|8k|Code|
|`deepseek-ai/deepseek-coder-6.7b-instruct`|DeepSeek|16k|Programmation|

\---

### `nim config` — Configuration

```batch
nim config                          # Afficher la config actuelle
nim config set model mistralai/mistral-large    # Changer le modèle
nim config set temperature 0.8      # Ajuster la créativité
nim config set max-tokens 2048      # Limite de tokens
nim config set stream false         # Désactiver le streaming
```

\---

## Configuration avancée

Le fichier de configuration se trouve dans :

* **Windows** : `%USERPROFILE%\\.nvidia-nim\\config.yaml`
* **Linux/macOS** : `\~/.nvidia-nim/config.yaml`

```yaml
# Exemple de config.yaml
api\_key: nvapi-votre-cle-ici
default\_model: meta/llama3-70b-instruct
max\_tokens: 1024
temperature: 0.7
stream: true
```

**Variable d'environnement (prioritaire sur le fichier) :**

```batch
set NVIDIA\_API\_KEY=nvapi-votre-cle
```

\---

## Architecture du projet

```
nvidia-nim-cli/
├── main.go                    # Point d'entrée
├── go.mod                     # Dépendances Go
├── build.bat                  # Script compilation Windows
├── Makefile                   # Script compilation Linux/Mac
│
├── cmd/                       # Commandes Cobra
│   ├── root.go                # Commande racine + init
│   ├── auth.go                # nim auth
│   ├── chat.go                # nim chat (mode interactif)
│   ├── ask.go                 # nim ask (one-shot)
│   ├── models.go              # nim models
│   └── config.go              # nim config
│
├── internal/
│   ├── api/
│   │   └── client.go          # Client HTTP NVIDIA NIM
│   ├── config/
│   │   └── config.go          # Gestion config (Viper)
│   └── ui/
│       └── styles.go          # Styles Lipgloss, spinners, bannière
│
└── pkg/
    └── models/
        └── models.go          # Catalogue des modèles NIM
```

\---

## Dépendances

|Librairie|Rôle|Version|
|-|-|-|
|`cobra`|Structure des commandes CLI|v1.8.0|
|`viper`|Gestion de la configuration|v1.18.2|
|`lipgloss`|Styles terminal (couleurs, bordures)|v0.10.0|
|`bubbletea`|TUI framework interactif|v0.26.1|
|`bubbles`|Composants UI (spinner, input)|v0.18.0|
|`resty`|Client HTTP REST|v2.12.0|
|`term`|Saisie masquée (clé API)|stdlib Go|

\---

## Obtenir votre clé API gratuite

1. Allez sur [https://build.nvidia.com/explore/discover](https://build.nvidia.com/explore/discover)
2. Cliquez sur **"Sign In"** (créez un compte si nécessaire)
3. Cliquez sur **"Get API Key"** en haut à droite
4. Générez une nouvelle clé et copiez-la (format : `nvapi-...`)
5. Lancez `nim auth` et collez votre clé

> \*\*Note :\*\* L'API NVIDIA NIM offre un crédit gratuit généreux. Consultez \[https://build.nvidia.com/pricing](https://build.nvidia.com) pour les détails.

\---

## Dépannage

**Erreur "go: not found"**
→ Go n'est pas installé ou pas dans le PATH. Redémarrez votre terminal après installation.

**Erreur "Clé API invalide"**
→ Vérifiez que votre clé commence par `nvapi-` et qu'elle est active sur build.nvidia.com

**Erreur réseau / timeout**
→ Vérifiez votre connexion internet. L'API NVIDIA nécessite une connexion HTTPS sortante.

**Le streaming ne fonctionne pas**
→ Désactivez temporairement : `nim chat --no-stream`

\---

## Licence

MIT License — Libre d'utilisation et de modification.

\---

*Développé par [Artleboss2](https://github.com/Artleboss2) en Go — Propulsé par NVIDIA AI Foundation Models*

