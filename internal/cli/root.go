// Package cli is missionctl's commands, output, and exit codes.
package cli

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/kmarsh/missionctl/internal/api"
	"github.com/kmarsh/missionctl/internal/config"
	"github.com/spf13/cobra"
)

// Exit codes, documented in the root help.
const (
	codeAPI    = 1
	codeUsage  = 64
	codeAuth   = 77
	codeConfig = 78
)

// exitError is a failure with its own exit code, for problems the API never sees.
type exitError struct {
	message string
	code    int
}

func (e *exitError) Error() string { return e.message }

func usageError(format string, args ...any) error {
	return &exitError{message: fmt.Sprintf(format, args...), code: codeUsage}
}

const longHelp = `missionctl works with your Mission Control projects and time entries.

Output is a table in a terminal and JSON when piped or with --json, so scripts
and agents get JSON by default. JSON is the API's own objects: list commands
print an array, the rest print one object. Errors go to stderr, as JSON in JSON
mode.

Projects can be given by name (case-insensitive; a unique partial name works
too) or by id. Names only match enabled projects, except with projects show
--all. Dates are YYYY-MM-DD, "today", or "yesterday". Durations are 1:30, :45,
a number of minutes (45), or hours and minutes with units: 45m, 10 min, 1.5h,
2 hrs, 1h30m, "1 hour 20 min".

The API key comes from $MISSIONCTL_API_KEY, or else api_key in
~/.mission_control.toml (or ~/.houston.toml), the Mission Control Mac app's
config.

Exit codes: 0 success, 1 API or network error, 64 bad usage or input,
77 API key rejected, 78 no API key or unreadable config.`

type app struct {
	version string
	json    bool
	now     func() time.Time
	client  *api.Client
}

// Run runs missionctl with args and returns its exit code.
func Run(version string, args []string) int {
	a := &app{version: version, now: time.Now}
	root := &cobra.Command{
		Use:           "missionctl",
		Short:         "Work with your Mission Control projects and time entries",
		Long:          longHelp,
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().BoolVar(&a.json, "json", false, "print JSON, even in a terminal")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError("%v", err) })
	root.AddCommand(a.projectsCommand(), a.timeCommand())
	root.SetArgs(args)

	err := root.Execute()
	if err == nil {
		return 0
	}
	return a.report(err)
}

// report prints err to stderr, as JSON in JSON mode, and picks its exit code.
func (a *app) report(err error) int {
	message, code, status := err.Error(), codeUsage, 0
	var exit *exitError
	var apiErr *api.Error
	var netErr *url.Error
	switch {
	case errors.As(err, &exit):
		code = exit.code
	case errors.As(err, &apiErr):
		message, status, code = apiErr.Message(), apiErr.Status, codeAPI
		if status == 401 {
			code = codeAuth
		}
	case errors.As(err, &netErr):
		message, code = "network error: "+netErr.Err.Error(), codeAPI
	}
	// Anything else came from parsing the command line.

	if a.output().json {
		body := map[string]any{"error": message}
		if status != 0 {
			body["status"] = status
		}
		writeJSON(os.Stderr, body)
	} else if status != 0 {
		fmt.Fprintf(os.Stderr, "missionctl: %s (HTTP %d)\n", message, status)
	} else {
		fmt.Fprintf(os.Stderr, "missionctl: %s\n", message)
	}
	return code
}

// do calls the API, keeping its errors distinct so report can pick exit codes.
func (a *app) do(method, path string, query url.Values, body any) (any, error) {
	if a.client == nil {
		key, err := config.APIKey()
		if err != nil {
			return nil, &exitError{
				message: err.Error() + ". Set $MISSIONCTL_API_KEY, or api_key in ~/.mission_control.toml",
				code:    codeConfig,
			}
		}
		a.client = api.New(key, a.version)
	}
	response, err := a.client.Do(method, path, query, body)
	var apiErr *api.Error
	var netErr *url.Error
	if err != nil && !errors.As(err, &apiErr) && !errors.As(err, &netErr) {
		return nil, &exitError{message: err.Error(), code: codeAPI}
	}
	return response, err
}

func (a *app) get(path string, query url.Values) (map[string]any, error) {
	response, err := a.do("GET", path, query, nil)
	return object(response), err
}

// all fetches every page of a paged list and joins them.
func (a *app) all(path, key string, query url.Values) ([]map[string]any, error) {
	var items []map[string]any
	for page := 1; ; page++ {
		query.Set("page", fmt.Sprint(page))
		query.Set("per_page", "500")
		response, err := a.get(path, query)
		if err != nil {
			return nil, err
		}
		items = append(items, objects(response[key])...)
		if page >= integer(object(response["pagination"]), "pages") {
			return items, nil
		}
	}
}
