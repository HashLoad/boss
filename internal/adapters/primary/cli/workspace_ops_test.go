//nolint:testpackage // exercises unexported command plumbing
package cli

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFlattenDetailRemovesBraces guards the rule that lets the PubPascal host
// tell a payload from an error: the host slices the captured output from the
// first brace to the last one, so an error detail carrying braces would be
// parsed as if it were the command's result.
func TestFlattenDetailRemovesBraces(t *testing.T) {
	got := flattenDetail("boom {\"error\":\"nope\"}\nsecond line")
	if strings.ContainsAny(got, "{}") {
		t.Errorf("flattenDetail kept a brace: %q", got)
	}
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("flattenDetail kept a newline: %q", got)
	}
}

// TestFlattenDetailTruncatesOnRuneBoundary checks the cap does not split a
// multi-byte rune, which would emit an invalid UTF-8 byte to the host.
func TestFlattenDetailTruncatesOnRuneBoundary(t *testing.T) {
	got := flattenDetail(strings.Repeat("ç", maxErrorDetailRunes+50))
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected a truncated detail, got %q", got)
	}
	if !strings.HasPrefix(got, strings.Repeat("ç", maxErrorDetailRunes)) {
		t.Error("truncation did not happen on a rune boundary")
	}
}

// TestPortalEndpointJoinsWithoutDoubleSlash covers a base URL saved with a
// trailing slash, which would otherwise produce //api/workspaces.
func TestPortalEndpointJoinsWithoutDoubleSlash(t *testing.T) {
	for _, base := range []string{"https://www.pubpascal.dev", "https://www.pubpascal.dev/"} {
		got := portalEndpoint(&PubPascalConfig{PortalBaseURL: base}, "/api/workspaces")
		if got != "https://www.pubpascal.dev/api/workspaces" {
			t.Errorf("base %q produced %q", base, got)
		}
	}
}

// TestWorkspaceListPayloadPreservesUnknownFields proves the pass-through: the
// desktop cockpit reads fields off each entry that this CLI never names, so
// decoding through a narrower struct would silently drop them.
func TestWorkspaceListPayloadPreservesUnknownFields(t *testing.T) {
	body := []byte(`{"workspaces":[{"id":"abc","name":"Janus","extra":42}]}`)

	var payload workspaceListPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	out, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != string(body) {
		t.Errorf("round trip changed the payload:\n got %s\nwant %s", out, body)
	}
}

// TestCatalogPayloadPreservesUnknownFields is the same guarantee for the
// catalog, whose entries carry slug/tier/score/stars/downloads.
func TestCatalogPayloadPreservesUnknownFields(t *testing.T) {
	body := []byte(`{"packages":[{"slug":"janus","name":"Janus","tier":"gold","score":95}]}`)

	var payload catalogPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	out, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != string(body) {
		t.Errorf("round trip changed the payload:\n got %s\nwant %s", out, body)
	}
}

// TestEmptyPayloadsMarshalAsArrays checks that "nothing found" is an empty
// array and not null: the desktop app calls .length on it.
func TestEmptyPayloadsMarshalAsArrays(t *testing.T) {
	cases := map[string]any{
		`{"workspaces":[]}`: workspaceListPayload{Workspaces: []json.RawMessage{}},
		`{"packages":[]}`:   catalogPayload{Packages: []json.RawMessage{}},
		`{"repos":[]}`:      workspaceDiffPayload{Repos: []workspaceRepoDiff{}},
	}

	for want, payload := range cases {
		out, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal %s: %v", want, err)
		}
		if string(out) != want {
			t.Errorf("got %s, want %s", out, want)
		}
	}
}

// TestWorkspaceDiffPayloadShape pins the field names the desktop overlay reads.
func TestWorkspaceDiffPayloadShape(t *testing.T) {
	out, err := json.Marshal(workspaceDiffPayload{
		Repos: []workspaceRepoDiff{{Name: "janus", Diff: "@@ -1 +1 @@\n-a\n+b\n"}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	const want = `{"repos":[{"name":"janus","diff":"@@ -1 +1 @@\n-a\n+b\n"}]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

// TestGetPortalJSONMapsStatusCodes checks that every failure the portal can
// return becomes a distinct, non-empty error instead of an empty payload.
func TestGetPortalJSONMapsStatusCodes(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"unauthorized", http.StatusUnauthorized, `{"error":"nope"}`, errPortalUnauthorized},
		{"forbidden", http.StatusForbidden, `{"error":"nope"}`, errPortalUnauthorized},
		{"rate limited", http.StatusTooManyRequests, `{"error":"Too many requests"}`, errPortalRateLimited},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			_, err := getPortalJSON(t.Context(), server.URL, "token")
			if !errors.Is(err, tc.want) {
				t.Errorf("got error %v, want %v", err, tc.want)
			}
		})
	}
}

// TestGetPortalJSONServerErrorDetailIsBraceFree makes sure a hostile or broken
// response body cannot smuggle a JSON object into the error message.
func TestGetPortalJSONServerErrorDetailIsBraceFree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html><style>body{color:red}</style></html>"))
	}))
	defer server.Close()

	_, err := getPortalJSON(t.Context(), server.URL, "")
	if err == nil {
		t.Fatal("expected an error for HTTP 502")
	}
	if strings.ContainsAny(err.Error(), "{}") {
		t.Errorf("error detail leaked a brace: %q", err)
	}
}

// TestGetPortalJSONSendsBearerOnlyWhenGiven covers the split between the
// authenticated workspace list and the public, login-free package catalog.
func TestGetPortalJSONSendsBearerOnlyWhenGiven(t *testing.T) {
	for _, token := range []string{"", "secret-token"} {
		var seen string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"packages":[]}`))
		}))

		if _, err := getPortalJSON(t.Context(), server.URL, token); err != nil {
			t.Fatalf("token %q: %v", token, err)
		}
		server.Close()

		want := ""
		if token != "" {
			want = "Bearer " + token
		}
		if seen != want {
			t.Errorf("token %q sent Authorization %q, want %q", token, seen, want)
		}
	}
}

// TestExcludeModulesPathspecTargetsModulesDir keeps the guard tied to the
// directory the workspace clone actually uses for dependencies.
func TestExcludeModulesPathspecTargetsModulesDir(t *testing.T) {
	if want := ":(exclude)" + modulesDirName; excludeModulesPathspec != want {
		t.Errorf("got %q, want %q", excludeModulesPathspec, want)
	}
}
