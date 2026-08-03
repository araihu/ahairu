package site

import "testing"

func TestProjectLifecycleStatusesMatchPublicMaturity(t *testing.T) {
	want := map[string]string{
		"Goshtoso":        "BETA",
		"Goshtoso Charts": "ALPHA",
		"Manja":           "WIP",
		"Pajé":            "WIP",
		"X-9":             "WIP",
	}

	for _, content := range Locales() {
		projects := append(append([]Project{}, content.Projects...), content.MoreProjects...)
		for _, project := range projects {
			if expected, tracked := want[project.Name]; tracked && project.Status != expected {
				t.Errorf("%s status in %s = %q, want %q", project.Name, content.Language, project.Status, expected)
			}
			if !containsProject(want, project.Name) && project.Status != "" {
				t.Errorf("%s status in %s = %q, want no public maturity label", project.Name, content.Language, project.Status)
			}
		}
	}
}

func containsProject(projects map[string]string, name string) bool {
	_, ok := projects[name]
	return ok
}
