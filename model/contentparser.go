package model

import (
	"bufio"
	"regexp"
	"strings"
)

var headerRegex *regexp.Regexp = regexp.MustCompile(`^#+\s*(.+)$`)
var definitionRegex *regexp.Regexp = regexp.MustCompile(`^\s*[-\*]?\s*(.+)$`)

func normalizeHeaders(sections map[string]string) map[string]string {
	//TODO load dict from file
	dict := map[string]string{
		"sito web": "website",
		"in breve": "short",
		"città":    "area",
	}
	result := make(map[string]string)
	for k, v := range sections {
		if newkey, ok := dict[k]; ok {
			result[newkey] = v
		} else {
			result[k] = v
		}
	}
	return result
}

// ParseContent parses some markdown-like content into its sections,
// returning a map[sectionName]content.
// If a section called "data" is met, its contents are interpreted "key:value"
// pairs, and added to the returned map.
func parseContent(content string, dataSectionName string) map[string]string {
	scanner := bufio.NewScanner(strings.NewReader(content))

	currentHeader := "short"
	dataMap := make(map[string]string)
	var currentContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		headerParts := headerRegex.FindStringSubmatch(line)
		if len(headerParts) == 2 {
			// Met a new header

			if currentContent.Len() > 0 {
				dataMap[currentHeader] = strings.TrimSpace(currentContent.String())
			}
			currentContent = strings.Builder{}
			currentHeader = strings.ToLower(strings.TrimSpace(headerParts[1]))
			continue
		} else {
			// Line is not an header

			if currentHeader == "data" || currentHeader == dataSectionName {
				// Processing a line inside of the "data" header
				defParts := definitionRegex.FindStringSubmatch(line)
				if len(defParts) == 2 {
					defParts = strings.SplitN(defParts[1], ":", 2)
					if len(defParts) == 2 {
						left := strings.ToLower(strings.TrimSpace(defParts[0]))
						dataMap[left] = strings.TrimSpace(defParts[1])
					}
				}
			} else {
				// Processing a line inside of another header
				currentContent.WriteString(line)
				currentContent.WriteString("\n")
			}
		}
	}

	if currentContent.Len() > 0 {
		dataMap[currentHeader] = strings.TrimSpace(currentContent.String())
	}
	return dataMap
}
