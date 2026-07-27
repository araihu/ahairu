package site

const brandAssetBase = "/assets/logos/"

const araihuMarkURL = brandAssetBase + "araihu-mark.svg"

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
	MarkURL     string
	MarkText    string
}

var contents = map[string]Content{
	"en": {
		Language: "en", Path: "/en/", Name: "Arai Hû", Tagline: "Independent software for work that matters.",
		Intro:         "Arai Hû is a small organization that builds and maintains durable, open software. In Guarani, arai hû means black or dark cloud.",
		ProjectsLabel: "Maintained projects", Meaning: "Arai Hû · black or dark cloud in Guarani.",
		Projects: projects("A Go UI library for server-rendered applications.", "OpenAPI documentation and publishing workbench.", "Durable workflows for code changes.", "Self-hosted monitoring control plane."),
	},
	"pt-br": {
		Language: "pt-BR", Path: "/pt-br/", Name: "Arai Hû", Tagline: "Software independente para trabalho que importa.",
		Intro:         "Arai Hû é uma pequena organização que cria e mantém software aberto e durável. Em guarani, arai hû significa nuvem preta ou escura.",
		ProjectsLabel: "Projetos mantidos", Meaning: "Arai Hû · nuvem preta ou escura em guarani.",
		Projects: projects("Biblioteca de UI Go para aplicações renderizadas no servidor.", "Documentação OpenAPI e ambiente de publicação.", "Workflows duráveis para mudanças de código.", "Plano de controle de monitoramento auto-hospedado."),
	},
	"es": {
		Language: "es", Path: "/es/", Name: "Arai Hû", Tagline: "Software independiente para trabajo importante.",
		Intro:         "Arai Hû es una organización pequeña que crea y mantiene software abierto y durable. En guaraní, arai hû significa nube negra u oscura.",
		ProjectsLabel: "Proyectos mantenidos", Meaning: "Arai Hû · nube negra u oscura en guaraní.",
		Projects: projects("Biblioteca de UI Go para aplicaciones renderizadas en servidor.", "Documentación OpenAPI y espacio de publicación.", "Flujos de trabajo durables para cambios de código.", "Plano de control de monitoreo autoalojado."),
	},
}

func projects(goshtoso, manja, paje, xisnove string) []Project {
	return []Project{
		{Name: "Goshtoso", Description: goshtoso, URL: "https://goshtoso.araihu.com", MarkURL: brandAssetBase + "goshtoso-mark.svg"},
		{Name: "Manja", Description: manja, URL: "https://manja.araihu.com", MarkURL: brandAssetBase + "manja-mark.svg"},
		{Name: "Pajé", Description: paje, URL: "https://paje.araihu.com", MarkURL: brandAssetBase + "paje-mark.svg"},
		{Name: "Xisnove", Description: xisnove, URL: "https://xisnove.dev", MarkText: "X-9"},
	}
}

// Locales returns all generated locales. English is the fallback.
func Locales() []Content { return []Content{contents["en"], contents["pt-br"], contents["es"]} }
