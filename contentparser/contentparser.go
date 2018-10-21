package contentparser

import (
	"bufio"
	"regexp"
	"strings"
)

var headerRegex *regexp.Regexp = regexp.MustCompile(`^#+\s*(.+)$`)
var definitionRegex *regexp.Regexp = regexp.MustCompile(`^\s*[-\*]?\s*(.+)$`)

func ParseContent(content string, dataHeader string) map[string]string {
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

			if currentHeader == "data" || currentHeader == dataHeader {
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
