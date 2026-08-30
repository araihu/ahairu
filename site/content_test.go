package site

import "testing"

func TestProjectLifecycleStatusesMatchPublicMaturity(t *testing.T) {
	want := map[string]string{
		"Goshtoso":            "BETA",
		"Goshtoso App Shells": "ALPHA",
		"Goshtoso Charts":     "ALPHA",
		"Manja":               "WIP",
		"Pajé":                "WIP",
		"X-9":                 "WIP",
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

func TestMargoAppearsInEveryLocale(t *testing.T) {
	for _, content := range Locales() {
		margo := content.MoreProjects[len(content.MoreProjects)-1]
		if margo.Name != "Margo" || margo.URL != "https://margo.araihu.com" || margo.MarkURL != "/assets/visuals/margo-icon.svg" {
			t.Errorf("Margo project in %s = %#v", content.Language, margo)
		}
		if margo.Category == "" || margo.Description == "" || margo.Status != "" {
			t.Errorf("Margo presentation in %s = %#v", content.Language, margo)
		}
	}
}
