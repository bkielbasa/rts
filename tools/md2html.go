package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	inputFile := flag.String("input", "", "Input markdown file")
	outputFile := flag.String("output", "", "Output HTML file (optional, defaults to input with .html extension)")
	title := flag.String("title", "", "HTML page title (optional, extracted from first heading)")
	theme := flag.String("theme", "dark", "Theme: dark or light")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: go run tools/md2html.go -input <file.md> [-output <file.html>] [-title <title>] [-theme dark|light]")
		fmt.Println("\nExample:")
		fmt.Println("  go run tools/md2html.go -input GAME_DESIGN.md")
		fmt.Println("  go run tools/md2html.go -input GAME_DESIGN_PL.md -output docs/design_pl.html -theme light")
		os.Exit(1)
	}

	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	if *outputFile == "" {
		ext := filepath.Ext(*inputFile)
		*outputFile = strings.TrimSuffix(*inputFile, ext) + ".html"
	}

	html := convertMarkdownToHTML(string(content), *title, *theme)

	err = os.WriteFile(*outputFile, []byte(html), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s\n", *outputFile)
}

func convertMarkdownToHTML(markdown, title, theme string) string {
	lines := strings.Split(markdown, "\n")
	var htmlLines []string
	inCodeBlock := false
	inTable := false
	inList := false
	tableHeaders := false

	if title == "" {
		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				title = strings.TrimPrefix(line, "# ")
				break
			}
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				htmlLines = append(htmlLines, "</code></pre>")
				inCodeBlock = false
			} else {
				lang := strings.TrimPrefix(line, "```")
				htmlLines = append(htmlLines, fmt.Sprintf("<pre><code class=\"language-%s\">", lang))
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			htmlLines = append(htmlLines, escapeHTML(line))
			continue
		}

		if strings.HasPrefix(line, "|") && strings.HasSuffix(strings.TrimSpace(line), "|") {
			if !inTable {
				htmlLines = append(htmlLines, "<div class=\"table-wrapper\"><table>")
				inTable = true
				tableHeaders = true
			}

			if strings.Contains(line, "---") && strings.Count(line, "|") > 1 {
				continue
			}

			cells := strings.Split(strings.Trim(line, "|"), "|")
			tag := "td"
			if tableHeaders {
				tag = "th"
				htmlLines = append(htmlLines, "<thead><tr>")
			} else {
				htmlLines = append(htmlLines, "<tr>")
			}

			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				cell = processInlineMarkdown(cell)
				htmlLines = append(htmlLines, fmt.Sprintf("<%s>%s</%s>", tag, cell, tag))
			}

			if tableHeaders {
				htmlLines = append(htmlLines, "</tr></thead><tbody>")
				tableHeaders = false
			} else {
				htmlLines = append(htmlLines, "</tr>")
			}
			continue
		} else if inTable {
			htmlLines = append(htmlLines, "</tbody></table></div>")
			inTable = false
		}

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			if !inList {
				htmlLines = append(htmlLines, "<ul>")
				inList = true
			}
			content := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			content = processInlineMarkdown(content)
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", content))
			continue
		} else if inList && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "  ") {
			htmlLines = append(htmlLines, "</ul>")
			inList = false
		}

		if matched, _ := regexp.MatchString(`^\d+\. `, line); matched {
			if !inList {
				htmlLines = append(htmlLines, "<ol>")
				inList = true
			}
			re := regexp.MustCompile(`^\d+\. `)
			content := re.ReplaceAllString(line, "")
			content = processInlineMarkdown(content)
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", content))
			continue
		}

		if strings.HasPrefix(line, "######") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "###### "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h6>%s</h6>", content))
		} else if strings.HasPrefix(line, "#####") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "##### "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h5>%s</h5>", content))
		} else if strings.HasPrefix(line, "####") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "#### "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h4>%s</h4>", content))
		} else if strings.HasPrefix(line, "###") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "### "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h3>%s</h3>", content))
		} else if strings.HasPrefix(line, "##") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "## "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h2 id=\"%s\">%s</h2>", slugify(content), content))
		} else if strings.HasPrefix(line, "#") {
			content := processInlineMarkdown(strings.TrimPrefix(line, "# "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h1>%s</h1>", content))
		} else if strings.TrimSpace(line) == "---" {
			htmlLines = append(htmlLines, "<hr>")
		} else if strings.TrimSpace(line) != "" {
			content := processInlineMarkdown(line)
			htmlLines = append(htmlLines, fmt.Sprintf("<p>%s</p>", content))
		}
	}

	if inTable {
		htmlLines = append(htmlLines, "</tbody></table></div>")
	}
	if inList {
		htmlLines = append(htmlLines, "</ul>")
	}
	if inCodeBlock {
		htmlLines = append(htmlLines, "</code></pre>")
	}

	toc := generateTOC(markdown)

	return generateFullHTML(strings.Join(htmlLines, "\n"), title, theme, toc)
}

func processInlineMarkdown(text string) string {
	boldItalic := regexp.MustCompile(`\*\*\*(.+?)\*\*\*`)
	text = boldItalic.ReplaceAllString(text, "<strong><em>$1</em></strong>")

	bold := regexp.MustCompile(`\*\*(.+?)\*\*`)
	text = bold.ReplaceAllString(text, "<strong>$1</strong>")

	italic := regexp.MustCompile(`\*(.+?)\*`)
	text = italic.ReplaceAllString(text, "<em>$1</em>")

	code := regexp.MustCompile("`([^`]+)`")
	text = code.ReplaceAllString(text, "<code>$1</code>")

	link := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = link.ReplaceAllString(text, "<a href=\"$2\">$1</a>")

	return text
}

func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

func slugify(text string) string {
	text = strings.ToLower(text)
	re := regexp.MustCompile(`[^a-z0-9ąćęłńóśźżа-яё]+`)
	text = re.ReplaceAllString(text, "-")
	text = strings.Trim(text, "-")
	return text
}

func generateTOC(markdown string) string {
	var toc []string
	scanner := bufio.NewScanner(strings.NewReader(markdown))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			title := strings.TrimPrefix(line, "## ")
			slug := slugify(title)
			toc = append(toc, fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", slug, title))
		}
	}

	if len(toc) == 0 {
		return ""
	}

	return fmt.Sprintf("<nav class=\"toc\"><h3>📑 Contents</h3><ul>%s</ul></nav>", strings.Join(toc, "\n"))
}

func generateFullHTML(content, title, theme, toc string) string {
	var bgColor, textColor, accentColor, cardBg, borderColor, codeBg, tableBg, tableHeaderBg string

	if theme == "dark" {
		bgColor = "#0a0e17"
		textColor = "#e0e6ed"
		accentColor = "#00d4ff"
		cardBg = "#151b28"
		borderColor = "#2a3547"
		codeBg = "#1a2234"
		tableBg = "#151b28"
		tableHeaderBg = "#1e2a3d"
	} else {
		bgColor = "#f5f7fa"
		textColor = "#1a1a2e"
		accentColor = "#0066cc"
		cardBg = "#ffffff"
		borderColor = "#d1d9e6"
		codeBg = "#f0f4f8"
		tableBg = "#ffffff"
		tableHeaderBg = "#e8edf4"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        :root {
            --bg-color: %s;
            --text-color: %s;
            --accent-color: %s;
            --card-bg: %s;
            --border-color: %s;
            --code-bg: %s;
            --table-bg: %s;
            --table-header-bg: %s;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
            background: var(--bg-color);
            color: var(--text-color);
            line-height: 1.7;
            padding: 2rem;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        h1 {
            font-size: 2.5rem;
            color: var(--accent-color);
            margin-bottom: 0.5rem;
            border-bottom: 3px solid var(--accent-color);
            padding-bottom: 0.5rem;
            text-shadow: 0 0 30px var(--accent-color);
        }

        h2 {
            font-size: 1.8rem;
            color: var(--accent-color);
            margin: 2rem 0 1rem;
            padding-bottom: 0.3rem;
            border-bottom: 1px solid var(--border-color);
        }

        h3 {
            font-size: 1.4rem;
            color: var(--text-color);
            margin: 1.5rem 0 0.8rem;
        }

        h4, h5, h6 {
            margin: 1rem 0 0.5rem;
        }

        p {
            margin-bottom: 1rem;
        }

        hr {
            border: none;
            border-top: 1px solid var(--border-color);
            margin: 2rem 0;
        }

        .toc {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 2rem;
        }

        .toc h3 {
            margin: 0 0 1rem 0;
            color: var(--accent-color);
        }

        .toc ul {
            list-style: none;
            columns: 2;
            gap: 2rem;
        }

        .toc li {
            margin-bottom: 0.5rem;
        }

        .toc a {
            color: var(--text-color);
            text-decoration: none;
            transition: color 0.2s;
        }

        .toc a:hover {
            color: var(--accent-color);
        }

        .table-wrapper {
            overflow-x: auto;
            margin: 1rem 0;
            border-radius: 8px;
            border: 1px solid var(--border-color);
        }

        table {
            width: 100%%;
            border-collapse: collapse;
            background: var(--table-bg);
        }

        th {
            background: var(--table-header-bg);
            font-weight: 600;
            text-align: left;
            padding: 0.8rem 1rem;
            border-bottom: 2px solid var(--accent-color);
        }

        td {
            padding: 0.6rem 1rem;
            border-bottom: 1px solid var(--border-color);
        }

        tr:hover td {
            background: var(--table-header-bg);
        }

        code {
            background: var(--code-bg);
            padding: 0.2rem 0.4rem;
            border-radius: 4px;
            font-family: 'Fira Code', 'Consolas', monospace;
            font-size: 0.9em;
        }

        pre {
            background: var(--code-bg);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 1rem;
            overflow-x: auto;
            margin: 1rem 0;
        }

        pre code {
            background: none;
            padding: 0;
        }

        strong {
            color: var(--accent-color);
            font-weight: 600;
        }

        ul, ol {
            margin: 1rem 0;
            padding-left: 2rem;
        }

        li {
            margin-bottom: 0.5rem;
        }

        a {
            color: var(--accent-color);
        }

        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }

            h1 {
                font-size: 1.8rem;
            }

            h2 {
                font-size: 1.4rem;
            }

            .toc ul {
                columns: 1;
            }

            table {
                font-size: 0.85rem;
            }

            th, td {
                padding: 0.5rem;
            }
        }

        @media print {
            body {
                background: white;
                color: black;
            }

            .toc {
                page-break-after: always;
            }

            h2 {
                page-break-before: always;
            }

            table {
                page-break-inside: avoid;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        %s
        %s
    </div>
</body>
</html>`, title, bgColor, textColor, accentColor, cardBg, borderColor, codeBg, tableBg, tableHeaderBg, toc, content)
}
