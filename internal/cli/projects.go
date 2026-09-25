package cli

import (
	"net/url"

	"github.com/kmarsh/missionctl/internal/match"
	"github.com/spf13/cobra"
)

func (a *app) projectsCommand() *cobra.Command {
	var all, disabled bool
	list := func(*cobra.Command, []string) error {
		query := url.Values{}
		if all {
			query.Set("all", "true")
		}
		if disabled {
			query.Set("disabled", "true")
		}
		response, err := a.get("projects.json", query)
		if err != nil {
			return err
		}
		projects := objects(response["projects"])

		out := a.output()
		if out.json {
			out.printJSON(projects)
			return nil
		}
		var rows [][]string
		for _, project := range projects {
			state := ""
			if !flag(project, "enabled", true) {
				state = "disabled"
			}
			rows = append(rows, []string{
				out.dot(str(project, "color")) + " " + str(project, "name"),
				str(project, "group"),
				state,
				str(project, "id"),
			})
		}
		out.table(rows)
		return nil
	}
	listFlags := func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().BoolVar(&all, "all", false, "include disabled projects")
		cmd.Flags().BoolVar(&disabled, "disabled", false, "only disabled projects")
		cmd.MarkFlagsMutuallyExclusive("all", "disabled")
		return cmd
	}

	projects := listFlags(&cobra.Command{
		Use:   "projects",
		Short: "List and look up projects",
		Long:  "List and look up projects. With no subcommand, lists enabled projects.",
		Args:  cobra.NoArgs,
		RunE:  list,
	})
	projects.AddCommand(
		listFlags(&cobra.Command{
			Use:   "list",
			Short: "List projects, enabled ones by default",
			Args:  cobra.NoArgs,
			RunE:  list,
		}),
		a.projectShowCommand(),
	)
	return projects
}

func (a *app) projectShowCommand() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "show PROJECT",
		Short: "Show one project, by name or id",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := a.projectID(args[0], all)
			if err != nil {
				return err
			}
			project, err := a.get("projects/"+id+".json", nil)
			if err != nil {
				return err
			}

			out := a.output()
			if out.json {
				out.printJSON(project)
				return nil
			}
			enabled := "yes"
			if !flag(project, "enabled", true) {
				enabled = "no"
			}
			out.table([][]string{
				{"Name", out.dot(str(project, "color")) + " " + str(project, "name")},
				{"Group", orDash(str(project, "group"))},
				{"Enabled", enabled},
				{"Color", orDash(str(project, "color"))},
				{"ID", id},
			})
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "match disabled projects' names too")
	return cmd
}

// projectID resolves a project name, or passes an id through. Names only match
// enabled projects unless includeDisabled.
func (a *app) projectID(query string, includeDisabled bool) (string, error) {
	if match.IsID(query) {
		return query, nil
	}
	params := url.Values{}
	if includeDisabled {
		params.Set("all", "true")
	}
	response, err := a.get("projects.json", params)
	if err != nil {
		return "", err
	}
	var candidates []match.Project
	for _, project := range objects(response["projects"]) {
		candidates = append(candidates, match.Project{
			ID:      str(project, "id"),
			Name:    str(project, "name"),
			Enabled: flag(project, "enabled", true),
		})
	}
	found, err := match.Find(query, candidates)
	if err != nil {
		return "", usageError("%v", err)
	}
	return found.ID, nil
}

func orDash(text string) string {
	if text == "" {
		return "–"
	}
	return text
}
