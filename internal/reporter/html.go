package reporter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/assets"
	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// GenerateStandaloneHTML bundles the React 19 visual studio and plan data into a self-contained HTML file.
func GenerateStandaloneHTML(plan *model.AnalyzedPlan) (string, error) {
	distSub, err := fs.Sub(assets.DistFS, "dist")
	if err != nil {
		return "", fmt.Errorf("failed to open embedded dist directory: %w", err)
	}

	indexBytes, err := fs.ReadFile(distSub, "index.html")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded index.html: %w", err)
	}

	indexHTML := string(indexBytes)

	// Read and inline CSS and JS assets
	var inlineCSS strings.Builder
	var inlineJS strings.Builder

	entries, err := fs.ReadDir(distSub, "assets")
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, ".css") {
				cssBytes, err := fs.ReadFile(distSub, "assets/"+name)
				if err == nil {
					inlineCSS.WriteString(fmt.Sprintf("<style>\n%s\n</style>\n", string(cssBytes)))
				}
			} else if strings.HasSuffix(name, ".js") {
				jsBytes, err := fs.ReadFile(distSub, "assets/"+name)
				if err == nil {
					inlineJS.WriteString(fmt.Sprintf("<script type=\"module\">\n%s\n</script>\n", string(jsBytes)))
				}
			}
		}
	}

	// Serialize plan to JSON
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("failed to serialize plan: %w", err)
	}

	// Script injecting preloaded plan into window
	preloadScript := fmt.Sprintf("<script>\nwindow.__TOFUFORGE_PRELOADED_PLAN__ = %s;\n</script>\n", string(planJSON))

	// Replace external asset tags with inlined bundles
	lines := strings.Split(indexHTML, "\n")
	var cleanedHTML strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<link rel=\"stylesheet\"") && strings.Contains(trimmed, "assets/") {
			continue
		}
		if strings.HasPrefix(trimmed, "<script type=\"module\"") && strings.Contains(trimmed, "assets/") {
			continue
		}
		cleanedHTML.WriteString(line + "\n")
	}

	finalHTML := cleanedHTML.String()

	// Inject preloadScript and CSS before </head>
	headReplacement := fmt.Sprintf("%s\n%s\n</head>", preloadScript, inlineCSS.String())
	finalHTML = strings.Replace(finalHTML, "</head>", headReplacement, 1)

	// Inject JS before </body>
	bodyReplacement := fmt.Sprintf("%s\n</body>", inlineJS.String())
	finalHTML = strings.Replace(finalHTML, "</body>", bodyReplacement, 1)

	return finalHTML, nil
}
