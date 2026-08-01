package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

// Names of the workspace sub-commands backed by the PubPascal portal.
const (
	subCmdNameList   = "list"
	subCmdNameSearch = "search"
)

// flagNameJSON switches a sub-command from its human-readable report to the
// machine payload consumed by the PubPascal desktop app and the RAD Studio
// (OTA) plugin. Both spawn this CLI and parse its standard output.
const flagNameJSON = "json"

// maxErrorDetailRunes caps how much of a portal or git error is echoed back.
const maxErrorDetailRunes = 200

// workspaceListPayload is the response of GET /api/workspaces.
//
// The entries are kept as raw JSON so every field the portal sends survives the
// round trip untouched: this command is a pass-through, and re-encoding through
// a narrower struct would silently drop any field added to the API later.
type workspaceListPayload struct {
	Workspaces []json.RawMessage `json:"workspaces"`
}

// catalogPayload is the response of GET /api/packages/catalog, kept raw for the
// same reason as workspaceListPayload -- the desktop cockpit reads slug, name,
// description, tier, score, stars, downloads and repository_url off each entry.
type catalogPayload struct {
	Packages []json.RawMessage `json:"packages"`
}

// workspaceSummary is the subset of a workspace entry the text report prints.
type workspaceSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Portal failures that carry no detail beyond the status code itself.
var (
	errPortalUnauthorized = errors.New("unauthorized, your portal auth token is invalid or expired")
	errPortalRateLimited  = errors.New("rate limited by the portal, try again in a moment")
)

// newWorkspaceListCmd builds 'boss workspace list'.
func newWorkspaceListCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   subCmdNameList,
		Short: "List the workspaces you own on the PubPascal portal",
		Long: "List the workspaces you own on the PubPascal portal.\n\n" +
			"With --json the portal payload is echoed verbatim on standard output, as " +
			"an object holding a \"workspaces\" array of id/name entries.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspaceList(cmd.Context(), asJSON)
		},
	}

	cmd.Flags().BoolVar(&asJSON, flagNameJSON, false, "print the portal payload as JSON on standard output")

	return cmd
}

// newWorkspaceSearchCmd builds 'boss workspace search'.
func newWorkspaceSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   subCmdNameSearch + " [query]",
		Short: "Search the public PubPascal package catalog",
		Long: "Search the public PubPascal package catalog.\n\n" +
			"The catalog is public, so no login is required. Without a query the portal " +
			"returns its curated front page. The result is always printed on standard " +
			"output as an object holding a \"packages\" array.",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := ""
			if len(args) == 1 {
				query = args[0]
			}
			runWorkspaceSearch(cmd.Context(), query)
		},
	}
}

// runWorkspaceList fetches the caller's workspaces from the portal.
//
// A token is mandatory here: /api/workspaces resolves the viewer from the
// bearer token and fail-softs to an empty list for anyone it cannot identify,
// so running without one would print "no workspaces" instead of "not logged
// in". The absent-token case is therefore rejected locally, before the call.
func runWorkspaceList(ctx context.Context, asJSON bool) {
	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal configuration: %s", flattenDetail(err.Error()))
	}

	if config.AuthToken == "" {
		msg.Die("❌ You must log in first. Run 'boss login --token <token>' with your portal token.")
	}

	body, err := getPortalJSON(ctx, portalEndpoint(config, "/api/workspaces"), config.AuthToken)
	if err != nil {
		msg.Die("❌ Failed to list workspaces: %s", err)
	}

	var payload workspaceListPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		msg.Die("❌ The portal did not return a workspace list: %s", flattenDetail(err.Error()))
	}
	if payload.Workspaces == nil {
		payload.Workspaces = []json.RawMessage{}
	}

	if asJSON {
		printJSONPayload(payload)

		return
	}

	printWorkspaceSummaries(payload.Workspaces)
}

// runWorkspaceSearch queries the public package catalog and echoes the payload.
//
// No credential is attached: /api/packages/catalog is public, and requiring a
// login to browse packages would lock the desktop cockpit's discover tab behind
// an account it does not need.
func runWorkspaceSearch(ctx context.Context, query string) {
	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal configuration: %s", flattenDetail(err.Error()))
	}

	endpoint := portalEndpoint(config, "/api/packages/catalog")
	if query = strings.TrimSpace(query); query != "" {
		endpoint += "?q=" + url.QueryEscape(query)
	}

	body, err := getPortalJSON(ctx, endpoint, "")
	if err != nil {
		msg.Die("❌ Failed to search the package catalog: %s", err)
	}

	var payload catalogPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		msg.Die("❌ The portal did not return a package catalog: %s", flattenDetail(err.Error()))
	}
	if payload.Packages == nil {
		payload.Packages = []json.RawMessage{}
	}

	printJSONPayload(payload)
}

// portalEndpoint joins the configured portal base URL with an API path.
func portalEndpoint(config *PubPascalConfig, path string) string {
	return strings.TrimSuffix(config.PortalBaseURL, "/") + path
}

// getPortalJSON performs a bounded GET against the portal and returns the body.
// An empty authToken sends no Authorization header, which is what the public
// endpoints expect.
func getPortalJSON(ctx context.Context, endpoint string, authToken string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not build the request: %w", err)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: portalRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := readPortalBody(resp)
	if err != nil {
		return nil, fmt.Errorf("could not read the portal response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, errPortalUnauthorized
	case http.StatusTooManyRequests:
		return nil, errPortalRateLimited
	default:
		return nil, fmt.Errorf("portal returned HTTP %d: %s", resp.StatusCode, portalErrorDetail(body))
	}
}

// printJSONPayload writes a payload as a single JSON line on standard output.
//
// Standard output is the contract: the PubPascal host merges the child's stdout
// and stderr and then recovers the payload by slicing from the first brace to
// the last brace of everything the process printed. Nothing else may be written
// after the payload, and no failure path may print a brace.
func printJSONPayload(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		msg.Die("❌ Failed to encode the JSON payload: %s", flattenDetail(err.Error()))
	}

	if _, err := fmt.Fprintln(os.Stdout, string(data)); err != nil {
		msg.Die("❌ Failed to write the JSON payload: %s", flattenDetail(err.Error()))
	}
}

// printWorkspaceSummaries prints the human-readable workspace report.
func printWorkspaceSummaries(entries []json.RawMessage) {
	if len(entries) == 0 {
		msg.Info("No workspaces found for this account.")

		return
	}

	msg.Info("Workspaces (%d):", len(entries))
	for _, entry := range entries {
		var summary workspaceSummary
		if err := json.Unmarshal(entry, &summary); err != nil {
			msg.Warn("  (skipped an entry the portal sent in an unexpected shape)")

			continue
		}
		msg.Info("  %s  %s", summary.ID, summary.Name)
	}
}

// portalErrorDetail reduces a portal error body to a short, brace-free string.
func portalErrorDetail(body []byte) string {
	return flattenDetail(portalErrorMessage(body))
}

// flattenDetail makes an arbitrary error string safe to print next to a JSON
// contract: braces are removed and the text is collapsed onto a single, capped
// line.
//
// The PubPascal host recovers a payload by slicing from the first brace to the
// last brace of the whole captured output, so an error body that carried braces
// -- a JSON blob, an HTML page with inline CSS -- would be picked up and parsed
// as if it were the command's result.
func flattenDetail(detail string) string {
	replacer := strings.NewReplacer("{", "", "}", "", "\r", " ", "\n", " ", "\t", " ")
	flattened := strings.TrimSpace(replacer.Replace(detail))

	runes := []rune(flattened)
	if len(runes) > maxErrorDetailRunes {
		return string(runes[:maxErrorDetailRunes]) + "..."
	}

	return flattened
}
