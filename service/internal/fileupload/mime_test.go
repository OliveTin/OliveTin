package fileupload

import "testing"

func TestMimeAllowed_plainWithCharset(t *testing.T) {
	allowed := []string{"text/plain"}
	if !mimeAllowed("text/plain; charset=utf-8", allowed) {
		t.Fatal("text/plain rule should allow text/plain with charset parameter")
	}
}

func TestMimeAllowed_exactPlain(t *testing.T) {
	allowed := []string{"text/plain"}
	if !mimeAllowed("text/plain", allowed) {
		t.Fatal("text/plain should match text/plain")
	}
}

func TestMimeAllowed_wildcard(t *testing.T) {
	allowed := []string{"text/*"}
	if !mimeAllowed("text/plain; charset=utf-8", allowed) {
		t.Fatal("text/* should allow text/plain with charset")
	}
}

func TestMimeAllowed_rejectOther(t *testing.T) {
	allowed := []string{"text/plain"}
	if mimeAllowed("application/octet-stream", allowed) {
		t.Fatal("application/octet-stream should not match text/plain")
	}
}

func TestMimeAllowed_ruleWithParams(t *testing.T) {
	allowed := []string{"text/plain; charset=utf-8"}
	if !mimeAllowed("text/plain; charset=utf-8", allowed) {
		t.Fatal("config with charset should still match")
	}
}
