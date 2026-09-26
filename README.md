# backlinks4hugo

> **Module Hugo & Hugo Blox pour prioriser les rétroliens (backlinks) dans les articles recommandés (« Sur le même sujet »).**

`backlinks4hugo` analyse automatiquement les liens internes de votre site statique Hugo, génère un index statique inversé (`data/backlinks.json`) et affiche en tête de la section **« Sur le même sujet »** les articles qui citent la page consultée.

---

## 🚀 Fonctionnalités

* **Indexation ultra-rapide en Go** : Analyse plus de 18 000 fichiers Markdown et résout les permaliens/alias en moins de 500 ms.
* **Zéro ralentissement de la compilation Hugo** : L'index statique est interrogé en temps constant $O(1)$ dans les templates.
* **Intégration Hugo Blox clé en main** : Fournit le partial `layouts/_partials/page_related.html` qui s'intègre nativement à Hugo Blox.
* **Repli transparent (Fallback)** : Si un article a moins de 5 rétroliens, les emplacements restants sont complétés automatiquement par le moteur de recommandation standard d'Hugo (tags et catégories).

---

## 📦 Installation

### 1. Importer le module Hugo

Dans la configuration de votre site (`config/_default/module.yaml` ou `hugo.yaml`) :

```yaml
module:
  imports:
    - path: github.com/drgoulu/backlinks4hugo
```

Et dans `go.mod` de votre site :

```go
require github.com/drgoulu/backlinks4hugo v0.0.0
```

*(En développement local, vous pouvez utiliser une directive `replace` dans `go.mod`)* :
```go
replace github.com/drgoulu/backlinks4hugo => ../backlinks4hugo
```

---

## 🛠️ Utilisation de l'outil CLI (Go)

L'outil en ligne de commande inclus dans le module génère le fichier `data/backlinks.json`.

### Exécution directe

```bash
go run github.com/drgoulu/backlinks4hugo
```

Ou en local depuis le dossier du module :
```bash
go run /chemin/vers/backlinks4hugo
```

### Options en ligne de commande

```bash
go run github.com/drgoulu/backlinks4hugo [OPTIONS]

Options :
  -content string
        Chemin vers le dossier contenant les fichiers Markdown (défaut : "content")
  -output string
        Chemin du fichier JSON de sortie (défaut : "data/backlinks.json")
  -domains string
        Domaines séparés par des virgules considérés comme des liens internes (défaut : "drgoulu.com,www.drgoulu.com")
  -quiet
        Désactive les messages de log
```

### Intégration dans `package.json`

Dans le `package.json` de votre site Hugo :

```json
{
  "scripts": {
    "backlinks": "go run github.com/drgoulu/backlinks4hugo",
    "dev": "pnpm run backlinks && hugo server --disableFastRender",
    "build": "pnpm run backlinks && hugo --gc --minify"
  }
}
```

---

## 📄 Licence

MIT © [Philippe Guglielmetti (Dr. Goulu)](https://drgoulu.com/)
