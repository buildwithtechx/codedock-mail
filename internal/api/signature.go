package api

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

const MaxSignatureLength = 100_000

var allowedTags = map[string]bool{
	"p":          true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"blockquote": true,
	"div":        true,
	"span":       true,
	"br":         true,
	"strong":     true,
	"b":          true,
	"em":         true,
	"i":          true,
	"u":          true,
	"s":          true,
	"ul":         true,
	"ol":         true,
	"li":         true,
	"a":          true,
	"img":        true,
}

var voidTags = map[string]bool{
	"br":  true,
	"img": true,
}

func SanitizeSignatureHTML(input string) (string, error) {
	if len(input) > MaxSignatureLength {
		return "", fmt.Errorf("signature exceeds maximum allowed length of %d bytes", MaxSignatureLength)
	}

	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var out bytes.Buffer

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}
			return "", tokenizer.Err()
		}

		switch tt {
		case html.TextToken:
			out.Write(tokenizer.Raw())

		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			tagName := strings.ToLower(token.Data)
			if !allowedTags[tagName] {
				continue
			}

			sanitizedAttrs := sanitizeAttributes(tagName, token.Attr)
			out.WriteString("<" + tagName)
			for _, attr := range sanitizedAttrs {
				out.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, html.EscapeString(attr.Val)))
			}
			if voidTags[tagName] {
				out.WriteString(" />")
			} else {
				out.WriteString(">")
			}

		case html.EndTagToken:
			token := tokenizer.Token()
			tagName := strings.ToLower(token.Data)
			if allowedTags[tagName] && !voidTags[tagName] {
				out.WriteString("</" + tagName + ">")
			}
		}
	}

	return out.String(), nil
}

func sanitizeAttributes(tag string, attrs []html.Attribute) []html.Attribute {
	var kept []html.Attribute
	for _, attr := range attrs {
		key := strings.ToLower(attr.Key)
		val := strings.TrimSpace(attr.Val)

		if tag == "a" {
			if key == "title" {
				kept = append(kept, html.Attribute{Key: "title", Val: val})
			} else if key == "href" {
				lowerVal := strings.ToLower(val)
				if strings.HasPrefix(lowerVal, "https://") || strings.HasPrefix(lowerVal, "http://") || strings.HasPrefix(lowerVal, "mailto:") {
					kept = append(kept, html.Attribute{Key: "href", Val: val})
				}
			}
		} else if tag == "img" {
			if key == "alt" || key == "width" || key == "height" {
				kept = append(kept, html.Attribute{Key: key, Val: val})
			} else if key == "src" {
				lowerVal := strings.ToLower(val)
				if strings.HasPrefix(lowerVal, "https://") || strings.HasPrefix(lowerVal, "http://") {
					kept = append(kept, html.Attribute{Key: "src", Val: val})
				}
			}
		}
	}
	return kept
}

func FormatReplyBody(replyBody, originalBody, inReplyTo string) string {
	reply := replyBody
	if strings.TrimSpace(originalBody) != "" && strings.TrimSpace(inReplyTo) != "" {
		return fmt.Sprintf("%s<blockquote>%s</blockquote>", reply, originalBody)
	}
	return reply
}
