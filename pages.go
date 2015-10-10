package main

import (
	"strings"
	"time"
	"github.com/microcosm-cc/bluemonday"
	"github.com/shurcooL/github_flavored_markdown"
)

// Renders a page
func RenderPage(template string, context map[string]string) string {
	content := templates[template]
	for key, value := range context {
		content = strings.Replace(content, "{"+key+"}", value, -1)
	}
	return content
}

// Render Page to Container
func RenderContainer(template string, context map[string]string) string {
	context[":*"] = RenderPage(template, context)
	return RenderPage(CONTAINER_TEMPLATE, context)
}

// Renders a template (without context)
func RenderTemplate(name string) string {
	result := templates[name]
	changed := true

	for changed {
		oldResult := result
		for key, value := range templates {
			if key == name {
				continue
			}

			result = strings.Replace(result, "{:" + key + "}", value, -1)
		}
		// check for changes
		changed = oldResult != result
	}
	return result
}

// Creates a container collection
func CreateContainerTemplate(template string) {
	/*container := templates[template]
	for key, c := range templates {
		if key == template {
			continue
		}

	}*/
}

// Renders the date to the standard format
func RenderTimestamp(date time.Time) string {
	return date.Format("Monday, 2. January 15:04")
}

// Renders the Markdown to sanitized HTML
func RenderMarkdown(markdown string) string {
	markdownBytes := []byte(markdown)
	htmlBytes := github_flavored_markdown.Markdown(markdownBytes)
	sanitizedBytes := bluemonday.UGCPolicy().SanitizeBytes(htmlBytes)
	return string(sanitizedBytes)
}
