package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	dateRe      = regexp.MustCompile(`(?m)^date:\s*['"]?(\d{4}-\d{2}-\d{2})`)
	fileDateRe  = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})`)
	slugRe      = regexp.MustCompile(`(?m)^slug:\s*['"]?([^'"\r\n]+)`)
	titleRe     = regexp.MustCompile(`(?m)^title:\s*['"]?([^'"\r\n]+)['"]?`)
	draftRe     = regexp.MustCompile(`(?m)^draft:\s*(?i:true)`)
	aliasesRe   = regexp.MustCompile(`(?m)^aliases:\s*\n((?:\s*-\s*[^\r\n]+\n)+)`)
	aliasItemRe = regexp.MustCompile(`(?m)^\s*-\s*['"]?([^'"\r\n]+)`)
	linkRe      = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

type BacklinkItem struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

type PageInfo struct {
	Permalink string
	Title     string
	Date      string
	Body      string
}

func extractFrontmatter(content string) (string, string) {
	if !strings.HasPrefix(content, "---") {
		return "", content
	}
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return "", content
	}
	return parts[1], parts[2]
}


var (
	nonAlphaNumRe = regexp.MustCompile(`[^a-z0-9-_]+`)
	multiHyphenRe = regexp.MustCompile(`-+`)
)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacements := map[rune]string{
		'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae",
		'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
		'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
		'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o", 'œ': "oe",
		'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
		'ý': "y", 'ÿ': "y",
		'ç': "c", 'ñ': "n",
	}
	var sb strings.Builder
	for _, r := range s {
		if rep, ok := replacements[r]; ok {
			sb.WriteString(rep)
		} else {
			sb.WriteRune(r)
		}
	}
	res := nonAlphaNumRe.ReplaceAllString(sb.String(), "-")
	res = multiHyphenRe.ReplaceAllString(res, "-")
	return strings.Trim(res, "-")
}

func normalizePath(p string) string {

	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	return p
}

func main() {
	contentDir := flag.String("content", "content", "Chemin vers le dossier contenant les fichiers Markdown")
	outputFile := flag.String("output", filepath.Join("data", "backlinks.json"), "Chemin du fichier JSON de sortie")
	domainsFlag := flag.String("domains", "drgoulu.com,www.drgoulu.com", "Domaines séparés par des virgules considérés comme internes")
	quiet := flag.Bool("quiet", false, "Désactiver les messages de progression")
	orderFlag := flag.String("order", "asc", "Tri des rétroliens: asc (chronologique, défaut) ou desc (antéchronologique)")
	flag.Parse()

	start := time.Now()

	allowedDomains := make(map[string]bool)
	for _, d := range strings.Split(*domainsFlag, ",") {
		d = strings.TrimSpace(strings.ToLower(d))
		if d != "" {
			allowedDomains[d] = true
		}
	}

	pages := make(map[string]PageInfo)     // canonical permalink -> PageInfo
	aliasToPerm := make(map[string]string) // alias path -> canonical permalink
	var fileList []string

	err := filepath.WalkDir(*contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors du parcours de %s: %v\n", *contentDir, err)
		os.Exit(1)
	}

	// 1. Première passe : indexer toutes les pages valides
	for _, fpath := range fileList {
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		content := string(data)
		fm, body := extractFrontmatter(content)
		if fm == "" {
			continue
		}

		if draftRe.MatchString(fm) {
			continue
		}

		var year, month, day, dateStr string
		if m := dateRe.FindStringSubmatch(fm); len(m) > 1 {
			dateStr = strings.Trim(m[1], " \"'")
			parts := strings.Split(dateStr, "-")
			if len(parts) == 3 {
				year, month, day = parts[0], parts[1], parts[2]
			}
		} else {
			fname := filepath.Base(fpath)
			if m := fileDateRe.FindStringSubmatch(fname); len(m) > 3 {
				year, month, day = m[1], m[2], m[3]
				dateStr = fmt.Sprintf("%s-%s-%s", year, month, day)
			}
		}

		if year == "" || month == "" || day == "" {
			continue
		}

		var slug string
		fname := filepath.Base(fpath)
		if m := slugRe.FindStringSubmatch(fm); len(m) > 1 {
			slug = strings.Trim(m[1], " \"'")
		}
		if slug == "" {
			nameNoExt := strings.TrimSuffix(fname, ".md")
			if fileDateRe.MatchString(nameNoExt) && len(nameNoExt) > 11 {
				slug = nameNoExt[11:]
			} else {
				slug = nameNoExt
			}
		}
		slug = slugify(slug)

		title := slug
		if m := titleRe.FindStringSubmatch(fm); len(m) > 1 {
			t := strings.Trim(m[1], " \"'")
			if t != "" {
				title = t
			}
		}

		permalink := fmt.Sprintf("/%s/%s/%s/%s/", year, month, day, slug)
		pages[permalink] = PageInfo{
			Permalink: permalink,
			Title:     title,
			Date:      dateStr,
			Body:      body,
		}

		if m := aliasesRe.FindStringSubmatch(fm); len(m) > 1 {
			for _, item := range aliasItemRe.FindAllStringSubmatch(m[1], -1) {
				if len(item) > 1 {
					rawAlias := normalizePath(item[1])
					if rawAlias != "" {
						aliasToPerm[rawAlias] = permalink
						if unquoted, err := url.PathUnescape(rawAlias); err == nil && unquoted != rawAlias {
							aliasToPerm[unquoted] = permalink
						}
					}
				}
			}
		}
	}

	resolveTarget := func(dest string) string {
		dest = strings.TrimSpace(dest)
		u, err := url.Parse(dest)
		if err != nil {
			return ""
		}

		if u.Host != "" && !allowedDomains[strings.ToLower(u.Host)] {
			return ""
		}

		path := normalizePath(u.Path)
		if path == "" {
			return ""
		}

		if _, ok := pages[path]; ok {
			return path
		}
		if unquoted, err := url.PathUnescape(path); err == nil {
			if _, ok := pages[unquoted]; ok {
				return unquoted
			}
			if target, ok := aliasToPerm[unquoted]; ok {
				return target
			}
		}
		if target, ok := aliasToPerm[path]; ok {
			return target
		}

		return ""
	}

	// 2. Seconde passe : extraire les liens et construire la table inversée
	backlinks := make(map[string][]BacklinkItem)

	for srcPerm, page := range pages {
		seenTargets := make(map[string]bool)
		matches := linkRe.FindAllStringSubmatch(page.Body, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			dest := m[2]
			target := resolveTarget(dest)
			if target != "" && target != srcPerm && !seenTargets[target] {
				seenTargets[target] = true
				backlinks[target] = append(backlinks[target], BacklinkItem{
					URL:   srcPerm,
					Title: page.Title,
					Date:  page.Date,
				})
			}
		}
	}

	// 3. Tri chronologique pour chaque cible (ascendant par défaut pour afficher les suites directes)
	isDesc := strings.ToLower(strings.TrimSpace(*orderFlag)) == "desc"
	for target := range backlinks {
		items := backlinks[target]
		sort.Slice(items, func(i, j int) bool {
			if isDesc {
				return items[i].Date > items[j].Date
			}
			return items[i].Date < items[j].Date
		})
		backlinks[target] = items
	}

	// 4. Sérialisation JSON et écriture conditionnelle
	newJSON, err := json.MarshalIndent(backlinks, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur sérialisation JSON: %v\n", err)
		os.Exit(1)
	}

	outDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur création dossier %s: %v\n", outDir, err)
		os.Exit(1)
	}

	changed := true
	if oldData, err := os.ReadFile(*outputFile); err == nil {
		if bytes.Equal(oldData, newJSON) {
			changed = false
		}
	}

	if changed {
		if err := os.WriteFile(*outputFile, newJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur écriture %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
	}

	elapsed := time.Since(start)
	if !*quiet {
		totalLinks := 0
		for _, v := range backlinks {
			totalLinks += len(v)
		}
		status := "mis à jour"
		if !changed {
			status = "inchangé (à jour)"
		}
		fmt.Printf("✅ backlinks4hugo: %d cibles avec rétroliens (%d liens totaux) %s en %v\n",
			len(backlinks), totalLinks, status, elapsed.Round(time.Millisecond))
	}
}
