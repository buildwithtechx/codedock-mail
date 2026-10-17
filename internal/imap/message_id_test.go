package imap

import (
	"testing"
)

func TestResolveMessageIDCandidates(t *testing.T) {
	t.Run("resolves literal and bracketed candidates", func(t *testing.T) {
		candidates := ResolveMessageIDCandidates("<msg-42@example.com>")
		expected := []string{"<msg-42@example.com>", "msg-42@example.com"}
		for _, exp := range expected {
			found := false
			for _, c := range candidates {
				if c == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("candidate %q missing from %v", exp, candidates)
			}
		}
	})

	t.Run("resolves legacy URL-encoded wrappers", func(t *testing.T) {
		candidates := ResolveMessageIDCandidates("%3C6044072238196767142@google.com%3E")
		found := false
		for _, c := range candidates {
			if c == "<6044072238196767142@google.com>" || c == "6044072238196767142@google.com" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected decoded candidate in %v", candidates)
		}
	})
}

func TestVerifyExactMessageID(t *testing.T) {
	t.Run("exact case sensitivity match", func(t *testing.T) {
		if !VerifyExactMessageID("<Invoice@example.com>", "Invoice@example.com") {
			t.Error("expected exact match for bracketed and unbracketed ID")
		}
		// Case difference must NOT match
		if VerifyExactMessageID("<Invoice@example.com>", "<invoice@example.com>") {
			t.Error("different case must not match")
		}
		// Substring or prefix must NOT match
		if VerifyExactMessageID("<invoice@example.com>", "<otherinvoice@example.com>") {
			t.Error("substring match must not match")
		}
	})

	t.Run("rejects empty or missing identity", func(t *testing.T) {
		emptyCases := []string{"", "<>", "%3C%3E", "   "}
		for _, ec := range emptyCases {
			if VerifyExactMessageID(ec, "<other@example.com>") {
				t.Errorf("expected empty identity %q to be rejected", ec)
			}
		}
	})
}

func TestVerifyMessageMutation(t *testing.T) {
	t.Run("fails when target message is not in matched set", func(t *testing.T) {
		err := VerifyMessageMutation("<invoice@example.com>", []string{"<other@example.com>"})
		if err == nil || err.Error() != "Message not found" {
			t.Fatalf("expected Message not found, got %v", err)
		}
	})

	t.Run("succeeds when target message matches exactly", func(t *testing.T) {
		err := VerifyMessageMutation("<invoice@example.com>", []string{"invoice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
