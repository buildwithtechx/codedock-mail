package api

import (
	"strings"
	"testing"
)

func TestSanitizeSignatureHTML(t *testing.T) {
	t.Run("preserves safe formatting and links", func(t *testing.T) {
		input := `<p><strong>Alex</strong><br /><a href="mailto:alex@example.com">Email</a><img src="https://example.com/logo.png" alt="Logo" /></p>`
		got, err := SanitizeSignatureHTML(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "<strong>Alex</strong>") {
			t.Errorf("expected strong tag preserved, got %s", got)
		}
		if !strings.Contains(got, `href="mailto:alex@example.com"`) {
			t.Errorf("expected mailto link preserved, got %s", got)
		}
		if !strings.Contains(got, `src="https://example.com/logo.png"`) {
			t.Errorf("expected img src preserved, got %s", got)
		}
	})

	t.Run("strips unsafe scripts, styles and data URIs", func(t *testing.T) {
		input := `<script>alert(1)</script><p onclick="alert(1)">Alex</p><img src="data:image/svg+xml,evil" onerror="alert(1)" /><a href="javascript:alert(1)">Link</a><iframe src="https://evil.test"></iframe><style>body{display:none}</style>`
		got, err := SanitizeSignatureHTML(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(got, "script") || strings.Contains(got, "onclick") || strings.Contains(got, "onerror") || strings.Contains(got, "javascript") || strings.Contains(got, "iframe") || strings.Contains(got, "style") {
			t.Errorf("expected all dangerous tags and attributes stripped, got %s", got)
		}
		if strings.Contains(got, "data:image") {
			t.Errorf("expected data:image src stripped, got %s", got)
		}
	})

	t.Run("bounds stored markup size", func(t *testing.T) {
		large := strings.Repeat("a", MaxSignatureLength+1)
		_, err := SanitizeSignatureHTML(large)
		if err == nil {
			t.Fatal("expected error for oversized signature")
		}
	})
}

func TestFormatReplyBody(t *testing.T) {
	t.Run("quotes original message when replying", func(t *testing.T) {
		reply := "<p>Reply</p><p>Signature</p>"
		orig := "<p>Original report</p>"
		inReplyTo := "<msg-123@example.com>"
		got := FormatReplyBody(reply, orig, inReplyTo)
		expected := "<p>Reply</p><p>Signature</p><blockquote><p>Original report</p></blockquote>"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("leaves message unquoted if not a reply", func(t *testing.T) {
		msg := "<p>Fresh email</p>"
		got := FormatReplyBody(msg, "", "")
		if got != msg {
			t.Errorf("expected unchanged message, got %q", got)
		}
	})
}
