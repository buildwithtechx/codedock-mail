package imap

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeMessageID(id string) string {
	raw := strings.TrimSpace(id)
	if raw == "" {
		return ""
	}

	decoded := raw
	if strings.Contains(raw, "%") {
		if unescaped, err := url.PathUnescape(raw); err == nil && unescaped != "" {
			decoded = strings.TrimSpace(unescaped)
		} else if unescaped, err := url.QueryUnescape(raw); err == nil && unescaped != "" {
			decoded = strings.TrimSpace(unescaped)
		}
	}

	if strings.HasPrefix(decoded, "<") {
		decoded = decoded[1:]
	}
	if strings.HasSuffix(decoded, ">") {
		decoded = decoded[:len(decoded)-1]
	}
	return strings.TrimSpace(decoded)
}

func ResolveMessageIDCandidates(id string) []string {
	raw := strings.TrimSpace(id)
	if raw == "" {
		return nil
	}

	seen := make(map[string]bool)
	candidates := make([]string, 0, 4)

	add := func(c string) {
		c = strings.TrimSpace(c)
		if c != "" && !seen[c] {
			seen[c] = true
			candidates = append(candidates, c)
		}
	}

	add(raw)

	if strings.Contains(raw, "%") {
		if unescaped, err := url.PathUnescape(raw); err == nil {
			add(unescaped)
		} else if unescaped, err := url.QueryUnescape(raw); err == nil {
			add(unescaped)
		}
	}

	norm := NormalizeMessageID(raw)
	if norm != "" {
		add(norm)
		add("<" + norm + ">")
	}

	return candidates
}

func VerifyExactMessageID(expectedID, candidateMessageID string) bool {
	normExpected := NormalizeMessageID(expectedID)
	normCandidate := NormalizeMessageID(candidateMessageID)

	if normExpected == "" || normCandidate == "" {
		return false
	}

	// Exact string equality — case-sensitive match
	return normExpected == normCandidate
}

func VerifyMessageMutation(expectedID string, matchedIDs []string) error {
	normExpected := NormalizeMessageID(expectedID)
	if normExpected == "" {
		return fmt.Errorf("message identity cannot be empty")
	}

	found := false
	for _, matched := range matchedIDs {
		if VerifyExactMessageID(normExpected, matched) {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Message not found")
	}

	return nil
}
