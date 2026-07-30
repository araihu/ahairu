package site

const brandIconBase = BrandAssetsPublicPrefix + "icons/brand/"

const araihuIconURL = brandIconBase + "araihu-icon-adaptive-transparent-optical.svg"

// HomeContent is one localized home page.
type HomeContent struct {
	Language               string
	Path                   string
	Name                   string
	Tagline                string
	Intro                  string
	ProjectsLabel          string
	Meaning                string
	SkipLabel              string
	PrimaryNavigationLabel string
	FooterBuild            string
	Projects               []Project
}

// Project is a maintained AraiHu project.
type Project struct {
	Name        string
	Description string
	URL         string
	MarkURL     string
	MarkText    string
}

var homes = map[string]HomeContent{
	"en": {
		Language: "en", Path: "/en/", Name: "Arai Hû", Tagline: "Open software for building, documenting, changing, and monitoring systems.",
		Intro:         "Arai Hû maintains open software for server-rendered interfaces, API documentation, code-change workflows, and self-hosted monitoring. In Guarani, arai hû means black or dark cloud.",
		ProjectsLabel: "Maintained projects", Meaning: "Arai Hû · black or dark cloud in Guarani.",
		SkipLabel:              "Skip to content",
		PrimaryNavigationLabel: "Primary navigation",
		FooterBuild:            "Built with Go and Goshtoso.",
		Projects:               projects("Go UI components for server-rendered web applications.", "OpenAPI documentation and a publishing workbench.", "Durable workflows for code changes.", "A self-hosted monitoring control plane."),
	},
	"pt-br": {
		Language: "pt-BR", Path: "/pt-br/", Name: "Arai Hû", Tagline: "Software aberto para construir, documentar, mudar e monitorar sistemas.",
		Intro:         "Arai Hû mantém software aberto para interfaces renderizadas no servidor, documentação de APIs, workflows de mudança de código e monitoramento auto-hospedado. Em guarani, arai hû significa nuvem preta ou escura.",
		ProjectsLabel: "Projetos mantidos", Meaning: "Arai Hû · nuvem preta ou escura em guarani.",
		SkipLabel:              "Pular para o conteúdo",
		PrimaryNavigationLabel: "Navegação principal",
		FooterBuild:            "Criado com Go e Goshtoso.",
		Projects:               projects("Componentes de UI Go para aplicações web renderizadas no servidor.", "Documentação OpenAPI e ambiente de publicação.", "Workflows duráveis para mudanças de código.", "Um plano de controle de monitoramento auto-hospedado."),
	},
	"es": {
		Language: "es", Path: "/es/", Name: "Arai Hû", Tagline: "Software abierto para construir, documentar, cambiar y supervisar sistemas.",
		Intro:         "Arai Hû mantiene software abierto para interfaces renderizadas en servidor, documentación de API, flujos de cambios de código y monitoreo autoalojado. En guaraní, arai hû significa nube negra u oscura.",
		ProjectsLabel: "Proyectos mantenidos", Meaning: "Arai Hû · nube negra u oscura en guaraní.",
		SkipLabel:              "Saltar al contenido",
		PrimaryNavigationLabel: "Navegación principal",
		FooterBuild:            "Creado con Go y Goshtoso.",
		Projects:               projects("Componentes de UI Go para aplicaciones web renderizadas en servidor.", "Documentación OpenAPI y espacio de publicación.", "Flujos de trabajo durables para cambios de código.", "Un plano de control de monitoreo autoalojado."),
	},
}

func projects(goshtoso, manja, paje, x9 string) []Project {
	return []Project{
		{Name: "Goshtoso", Description: goshtoso, URL: "https://goshtoso.araihu.com", MarkURL: brandIconBase + "goshtoso-icon-adaptive-transparent-optical.svg"},
		{Name: "Manja", Description: manja, URL: "https://manja.araihu.com", MarkURL: brandIconBase + "manja-icon-adaptive-transparent-optical.svg"},
		{Name: "Pajé", Description: paje, URL: "https://paje.araihu.com", MarkURL: brandIconBase + "paje-icon-adaptive-transparent-optical.svg"},
		{Name: "X-9", Description: x9, URL: "https://x9.araihu.com", MarkURL: brandIconBase + "x9-icon-adaptive-transparent-optical.svg"},
	}
}

func homeContent(localeKey string) HomeContent { return homes[localeKey] }

// Locales returns all localized home pages. English is the fallback.
func Locales() []HomeContent { return []HomeContent{homes["en"], homes["pt-br"], homes["es"]} }
