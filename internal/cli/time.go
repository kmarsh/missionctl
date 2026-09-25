package cli

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/kmarsh/missionctl/internal/dates"
	"github.com/kmarsh/missionctl/internal/duration"
	"github.com/spf13/cobra"
)

const durationHelp = `how long: 1:30, :45, 45 (minutes), 45m, 1.5h, 1h30m, or "1 hour 20 min"`

func (a *app) timeCommand() *cobra.Command {
	var from, to, project string
	var today bool
	list := func(*cobra.Command, []string) error {
		query := url.Values{}
		now := a.now()
		switch {
		case today:
			date, _ := dates.Resolve("today", now)
			query.Set("from", date)
			query.Set("to", date)
		case from == "" && to == "":
			query.Set("from", dates.StartOfWeek(now))
		default:
			for name, value := range map[string]string{"from": from, "to": to} {
				if value == "" {
					continue
				}
				date, err := dates.Resolve(value, now)
				if err != nil {
					return usageError("%v", err)
				}
				query.Set(name, date)
			}
		}
		if project != "" {
			id, err := a.projectID(project, false)
			if err != nil {
				return err
			}
			query.Set("project_id", id)
		}
		entries, err := a.all("time_entries.json", "time_entries", query)
		if err != nil {
			return err
		}

		out := a.output()
		if out.json {
			out.printJSON(entries)
			return nil
		}
		if len(entries) == 0 {
			fmt.Println("No time entries")
			return nil
		}
		var rows [][]string
		minutes := 0
		for _, entry := range entries {
			rows = append(rows, a.entryRow(out, entry))
			minutes += integer(entry, "minutes")
		}
		out.table(rows)
		noun := "entries"
		if len(entries) == 1 {
			noun = "entry"
		}
		fmt.Printf("\nTotal %s in %d %s\n", duration.Format(minutes), len(entries), noun)
		return nil
	}
	listFlags := func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().StringVar(&from, "from", "", "first day to include")
		cmd.Flags().StringVar(&to, "to", "", "last day to include")
		cmd.Flags().BoolVar(&today, "today", false, "only today's entries")
		cmd.Flags().StringVarP(&project, "project", "p", "", "only entries for this project")
		cmd.MarkFlagsMutuallyExclusive("today", "from")
		cmd.MarkFlagsMutuallyExclusive("today", "to")
		return cmd
	}

	timeCmd := listFlags(&cobra.Command{
		Use:   "time",
		Short: "List, log, edit, and delete time entries",
		Long:  "List, log, edit, and delete time entries. With no subcommand, lists this week's entries.",
		Args:  cobra.NoArgs,
		RunE:  list,
	})
	timeCmd.AddCommand(
		listFlags(&cobra.Command{
			Use:   "list",
			Short: "List time entries, newest first",
			Long:  "List time entries, newest first. With no dates, lists this week's entries (from Monday).",
			Args:  cobra.NoArgs,
			RunE:  list,
		}),
		a.timeShowCommand(),
		a.timeLogCommand(),
		a.timeEditCommand(),
		a.timeDeleteCommand(),
	)
	return timeCmd
}

func (a *app) timeShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show ID",
		Short: "Show one time entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			entry, err := a.get("time_entries/"+args[0]+".json", nil)
			if err != nil {
				return err
			}
			out := a.output()
			if out.json {
				out.printJSON(entry)
				return nil
			}
			a.printEntry(out, entry)
			return nil
		},
	}
}

func (a *app) timeLogCommand() *cobra.Command {
	var length, project, date string
	var noBillable bool
	cmd := &cobra.Command{
		Use:   `log "DESCRIPTION" -d DURATION`,
		Short: "Log a time entry",
		Example: `  missionctl time log "Fixed the login form" --project Apollo --duration 1:30
  missionctl time log "Standup" -p "Gemini Mobile" -d 15 --date yesterday
  missionctl time log "Client call" -p Apollo -d "1 hour 20 min"`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			minutes, err := minutesFrom(length)
			if err != nil {
				return err
			}
			day, err := dates.Resolve(date, a.now())
			if err != nil {
				return usageError("%v", err)
			}
			body := map[string]any{"date": day, "minutes": minutes, "description": args[0], "billable": !noBillable}
			if project != "" {
				if body["project_id"], err = a.projectID(project, false); err != nil {
					return err
				}
			}
			response, err := a.do("POST", "time_entries.json", nil, body)
			if err != nil {
				return err
			}
			return a.reportEntry("Logged", object(response))
		},
	}
	cmd.Flags().StringVarP(&length, "duration", "d", "", durationHelp)
	cmd.Flags().StringVarP(&project, "project", "p", "", "the project's name or id")
	cmd.Flags().StringVar(&date, "date", "today", "the day worked")
	cmd.Flags().BoolVar(&noBillable, "no-billable", false, "log the time as not billable")
	_ = cmd.MarkFlagRequired("duration")
	return cmd
}

func (a *app) timeEditCommand() *cobra.Command {
	var length, description, project, date string
	var billable, noBillable bool
	cmd := &cobra.Command{
		Use:   "edit ID",
		Short: "Change a time entry; only the options given are changed",
		Example: `  missionctl time edit ENTRY_ID --duration "2 hrs"
  missionctl time edit ENTRY_ID --no-billable -m "Call with the client"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := cmd.Flags().Changed
			body := map[string]any{}
			if changed("duration") {
				minutes, err := minutesFrom(length)
				if err != nil {
					return err
				}
				body["minutes"] = minutes
			}
			if changed("description") {
				body["description"] = description
			}
			if changed("date") {
				day, err := dates.Resolve(date, a.now())
				if err != nil {
					return usageError("%v", err)
				}
				body["date"] = day
			}
			if billable || noBillable {
				body["billable"] = billable
			}
			if changed("project") {
				id, err := a.projectID(project, false)
				if err != nil {
					return err
				}
				body["project_id"] = id
			}
			if len(body) == 0 {
				return usageError("nothing to change; give at least one option")
			}
			response, err := a.do("PATCH", "time_entries/"+args[0]+".json", nil, body)
			if err != nil {
				return err
			}
			return a.reportEntry("Updated", object(response))
		},
	}
	cmd.Flags().StringVarP(&length, "duration", "d", "", durationHelp)
	cmd.Flags().StringVarP(&description, "description", "m", "", "what the time was spent on, in Markdown")
	cmd.Flags().StringVarP(&project, "project", "p", "", "the project's name or id")
	cmd.Flags().StringVar(&date, "date", "", "the day worked")
	cmd.Flags().BoolVar(&billable, "billable", false, "mark the time billable")
	cmd.Flags().BoolVar(&noBillable, "no-billable", false, "mark the time not billable")
	cmd.MarkFlagsMutuallyExclusive("billable", "no-billable")
	return cmd
}

func (a *app) timeDeleteCommand() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a time entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if !yes && !isTerminal(os.Stdin) {
				return usageError("pass --yes to delete without a prompt")
			}
			path := "time_entries/" + args[0] + ".json"
			entry, err := a.get(path, nil)
			if err != nil {
				return err
			}
			out := a.output()
			if !yes {
				invoiced := ""
				if str(entry, "invoice_id") != "" {
					invoiced = " It's already been invoiced."
				}
				fmt.Printf("Delete %s?%s [y/N] ", a.oneLine(out, entry), invoiced)
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Println("Not deleted")
					return nil
				}
			}
			if _, err := a.do("DELETE", path, nil, nil); err != nil {
				return err
			}
			return a.reportEntry("Deleted", entry)
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "delete without asking; required when not run from a terminal")
	return cmd
}

// minutesFrom reads a duration here rather than leaving it to the server, which
// accepts fewer forms.
func minutesFrom(text string) (int, error) {
	minutes, ok := duration.Minutes(text)
	if !ok {
		return 0, usageError("can't read duration %q; try 1:30, 45, 45m, 1.5h, or 1h30m", text)
	}
	return minutes, nil
}

// reportEntry prints the entry as JSON, or as "<verb> 1:30 on Fri, Sep 25 to ● Apollo: ...".
func (a *app) reportEntry(verb string, entry map[string]any) error {
	out := a.output()
	if out.json {
		out.printJSON(entry)
	} else {
		fmt.Println(verb + " " + a.oneLine(out, entry))
	}
	return nil
}

func (a *app) projectName(out output, entry map[string]any) string {
	project := object(entry["project"])
	if project == nil {
		return "No project"
	}
	return out.dot(str(project, "color")) + " " + str(project, "name")
}

func entryDuration(entry map[string]any) string {
	if text := str(entry, "duration"); text != "" {
		return text
	}
	return duration.Format(integer(entry, "minutes"))
}

func (a *app) entryRow(out output, entry map[string]any) []string {
	return []string{
		dates.Humanize(str(entry, "date"), a.now()),
		entryDuration(entry),
		a.projectName(out, entry),
		summary(str(entry, "description"), 60),
		str(entry, "id"),
	}
}

// oneLine is "1:30 on Fri, Sep 25 to ● Apollo: Fixed the login form (id)".
func (a *app) oneLine(out output, entry map[string]any) string {
	line := fmt.Sprintf("%s on %s to %s", entryDuration(entry), dates.Humanize(str(entry, "date"), a.now()), a.projectName(out, entry))
	if description := summary(str(entry, "description"), 60); description != "" {
		line += ": " + description
	}
	return line + " (" + str(entry, "id") + ")"
}

func (a *app) printEntry(out output, entry map[string]any) {
	yesNo := map[bool]string{true: "yes", false: "no"}
	out.table([][]string{
		{"Date", dates.Humanize(str(entry, "date"), a.now())},
		{"Duration", entryDuration(entry)},
		{"Project", a.projectName(out, entry)},
		{"Billable", yesNo[flag(entry, "billable", true)]},
		{"Invoiced", yesNo[str(entry, "invoice_id") != ""]},
		{"ID", str(entry, "id")},
	})
	if description := str(entry, "description"); description != "" {
		fmt.Println("\n" + description)
	}
}
