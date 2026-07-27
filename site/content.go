package site

// Content is one localized version of the organization site.
type Content struct {
	Language      string
	Path          string
	Name          string
	Tagline       string
	Intro         string
	ProjectsLabel string
	Meaning       string
	Projects      []Project
}

// Project is a maintained AraiHu project.
type Project struct {
	Name        string
	Description string
	URL         string
}

var contents = map[string]Content{
	"en": {
		Language: "en", Path: "/en/", Name: "Arai Hu", Tagline: "Independent software for work that matters.",
		Intro:         "Arai Hu is a small organization that builds and maintains durable, open software. Its name is a free interpretation of Tupi: black cloud.",
		ProjectsLabel: "Maintained projects", Meaning: "Arai Hu · black cloud, in free Tupi translation and interpretation.",
		Projects: projects("A Go UI library for server-rendered applications.", "OpenAPI documentation and publishing workbench.", "Durable workflows for code changes.", "Self-hosted monitoring control plane."),
	},
	"pt-br": {
		Language: "pt-BR", Path: "/pt-br/", Name: "Arai Hu", Tagline: "Software independente para trabalho que importa.",
		Intro:         "Arai Hu é uma pequena organização que cria e mantém software aberto e durável. Seu nome é uma tradução e interpretação livre do Tupi: nuvem negra.",
		ProjectsLabel: "Projetos mantidos", Meaning: "Arai Hu · nuvem negra, em tradução e interpretação livre do Tupi.",
		Projects: projects("Biblioteca de UI Go para aplicações renderizadas no servidor.", "Documentação OpenAPI e ambiente de publicação.", "Workflows duráveis para mudanças de código.", "Plano de controle de monitoramento auto-hospedado."),
	},
	"es": {
		Language: "es", Path: "/es/", Name: "Arai Hu", Tagline: "Software independiente para trabajo importante.",
		Intro:         "Arai Hu es una organización pequeña que crea y mantiene software abierto y durable. Su nombre es una traducción e interpretación libre del tupí: nube negra.",
		ProjectsLabel: "Proyectos mantenidos", Meaning: "Arai Hu · nube negra, en traducción e interpretación libre del tupí.",
		Projects: projects("Biblioteca de UI Go para aplicaciones renderizadas en servidor.", "Documentación OpenAPI y espacio de publicación.", "Flujos de trabajo durables para cambios de código.", "Plano de control de monitoreo autoalojado."),
	},
}

func projects(goshtoso, manja, paje, xisnove string) []Project {
	return []Project{
		{Name: "Goshtoso", Description: goshtoso, URL: "https://goshtoso.araihu.com"},
		{Name: "Manja", Description: manja, URL: "https://manja.araihu.com"},
		{Name: "Pajé", Description: paje, URL: "https://paje.araihu.com"},
		{Name: "Xisnove", Description: xisnove, URL: "https://xisnove.dev"},
	}
}

// Locales returns all generated locales. English is the fallback.
func Locales() []Content { return []Content{contents["en"], contents["pt-br"], contents["es"]} }
