package mapper

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	reCodeBlock      = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+)?\n?(.*?)```")
	reInlineCode     = regexp.MustCompile("(?s)`([^`\n]+)`")
	reBoldAsterisk   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reBoldUnderscore = regexp.MustCompile(`__([^_]+)__`)
	reBulletList     = regexp.MustCompile(`(?m)^\*\s+`)
	reTgTag          = regexp.MustCompile(`(?i)</?(?:b|strong|i|em|u|s|strike|del|code|pre|a)(?:\s+href="[^"]*")?>`)
)

func FormatTelegramHTML(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}

	var codeBlocks []string
	text := reCodeBlock.ReplaceAllStringFunc(input, func(m string) string {
		sub := reCodeBlock.FindStringSubmatch(m)
		if len(sub) > 1 {
			codeBlocks = append(codeBlocks, html.EscapeString(strings.Trim(sub[1], "\r\n")))
			return fmt.Sprintf("___CODE_BLOCK_%d___", len(codeBlocks)-1)
		}
		return m
	})

	var inlineCodes []string
	text = reInlineCode.ReplaceAllStringFunc(text, func(m string) string {
		sub := reInlineCode.FindStringSubmatch(m)
		if len(sub) > 1 {
			inlineCodes = append(inlineCodes, html.EscapeString(sub[1]))
			return fmt.Sprintf("___INLINE_CODE_%d___", len(inlineCodes)-1)
		}
		return m
	})

	var preservedTags []string
	text = reTgTag.ReplaceAllStringFunc(text, func(m string) string {
		preservedTags = append(preservedTags, m)
		return fmt.Sprintf("___TG_TAG_%d___", len(preservedTags)-1)
	})

	text = html.EscapeString(text)

	for i, tag := range preservedTags {
		placeholder := fmt.Sprintf("___TG_TAG_%d___", i)
		text = strings.Replace(text, placeholder, tag, 1)
	}

	text = reBoldAsterisk.ReplaceAllString(text, "<b>$1</b>")
	text = reBoldUnderscore.ReplaceAllString(text, "<b>$1</b>")
	text = reBulletList.ReplaceAllString(text, "• ")

	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		text = strings.Replace(text, placeholder, "<pre>"+block+"</pre>", 1)
	}

	for i, code := range inlineCodes {
		placeholder := fmt.Sprintf("___INLINE_CODE_%d___", i)
		text = strings.Replace(text, placeholder, "<code>"+code+"</code>", 1)
	}

	return text
}
