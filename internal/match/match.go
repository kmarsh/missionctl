// Package match finds the project a person means by name.
package match

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Project struct {
	ID      string
	Name    string
	Enabled bool
}

var uuid = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsID reports whether text is a project id rather than a name.
func IsID(text string) bool {
	return uuid.MatchString(text)
}

// Find returns an exact name match ignoring case, or else the one project whose
// name contains query, so "apollo" finds "Apollo" even alongside "Apollo Guidance". Partial
// matches look at enabled projects first, so an old disabled one doesn't get in the way.
func Find(query string, projects []Project) (Project, error) {
	needle := strings.ToLower(query)
	for _, project := range projects {
		if strings.ToLower(project.Name) == needle {
			return project, nil
		}
	}
	var matching, enabled []Project
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), needle) {
			matching = append(matching, project)
			if project.Enabled {
				enabled = append(enabled, project)
			}
		}
	}
	if len(enabled) > 0 {
		matching = enabled
	}
	switch len(matching) {
	case 1:
		return matching[0], nil
	case 0:
		return Project{}, fmt.Errorf("no project matches %q", query)
	}
	names := make([]string, len(matching))
	for i, project := range matching {
		names[i] = project.Name
	}
	sort.Strings(names)
	return Project{}, fmt.Errorf("%q matches several projects: %s", query, strings.Join(names, ", "))
}
