# backlinks4hugo

> **Hugo & Hugo Blox module to prioritize backlinks in recommended articles ("Related Content").**

`backlinks4hugo` automatically analyzes internal links across your Hugo static site, generates a static inverted index (`data/backlinks.json`), and features articles that reference the current page at the top of the **"Related Content"** section.

---

## 🚀 Features

* **Blazing-fast Go indexer**: Scans over 18,000 Markdown files and resolves canonical permalinks and aliases in under 500 ms.
* **Zero Hugo build slowdown**: The static index is queried in $O(1)$ constant time inside templates, introducing no overhead to compilation.
* **Turnkey Hugo Blox integration**: Includes a drop-in `layouts/_partials/page_related.html` partial that overrides Hugo Blox's default template.
* **Chronological Prioritization**: Backlinks are sorted chronologically (`asc` by default) so direct sequels and immediate follow-up articles appear first.
* **Seamless fallback**: If an article has fewer backlinks than the limit (default 5, configurable via `params.backlinks.limit`), remaining slots are filled by Hugo's standard taxonomy-based recommendation engine (tags and categories).

---

## 📦 Installation

### 1. Import the Hugo Module

In your site configuration (`config/_default/module.yaml` or `hugo.yaml`):

```yaml
module:
  imports:
    - path: github.com/drgoulu/backlinks4hugo
```

And in your site's `go.mod`:

```go
require github.com/drgoulu/backlinks4hugo v0.0.0
```

*(For local development, you can use a `replace` directive in `go.mod`)*:
```go
replace github.com/drgoulu/backlinks4hugo => ../backlinks4hugo
```

---

## 🛠️ CLI Tool Usage (Go)

The command-line tool included in the module generates the `data/backlinks.json` index.

### Direct Execution

```bash
go run github.com/drgoulu/backlinks4hugo
```

Or locally from the module directory:
```bash
go run /path/to/backlinks4hugo
```

### Command-Line Options

```bash
go run github.com/drgoulu/backlinks4hugo [OPTIONS]

Options:
  -content string
        Path to the folder containing Markdown content files (default: "content")
  -output string
        Path to the output JSON file (default: "data/backlinks.json")
  -domains string
        Comma-separated domains treated as internal links (default: "drgoulu.com,www.drgoulu.com")
  -order string
        Sorting order of backlinks: "asc" (chronological, oldest first, default) or "desc" (reverse-chronological)
  -quiet
        Suppress status and log messages
```

### Integration into `package.json`

In your Hugo site's `package.json`:

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

## 📄 License

MIT © [Philippe Guglielmetti (Dr. Goulu)](https://drgoulu.com/)
