package site

const brandAssetBase = "/assets/logos/"
const brandAssetRevision = "?rev=a8a9647a"

const araihuIconURL = brandAssetBase + "araihu-icon-transparent.svg" + brandAssetRevision

// Content is one localized version of the organization site.
type Content struct {
	Language               string
	Path                   string
	Name                   string
	Tagline                string
	Intro                  string
	ProjectsLabel          string
	ProjectsKicker         string
	ProjectsIntro          string
	ProjectAction          string
	OpenSourceLabel        string
	NavHome                string
	NavApps                string
	NavLibs                string
	NavBlog                string
	MenuLabel              string
	WorkInProgress         string
	FeaturedLabel          string
	FeaturedTitle          string
	FeaturedDescription    string
	FeaturedAction         string
	LibrariesLabel         string
	AppsLabel              string
	MoreLabel              string
	MoreIntro              string
	MissionTitle           string
	MissionCopy            string
	MailingTitle           string
	MailingCopy            string
	MailingPlaceholder     string
	MailingAction          string
	ChatTitle              string
	ChatCopy               string
	ChatAction             string
	Meaning                string
	SkipLabel              string
	PrimaryNavigationLabel string
	FooterBuiltWith        string
	FooterConjunction      string
	HeartChartLabel        string
	Projects               []Project
	MoreProjects           []Project
}

// Project is a maintained AraiHu project.
type Project struct {
	Name        string
	Category    string
	Description string
	URL         string
	MarkURL     string
	MarkText    string
	Status      string
}

var contents = map[string]Content{
	"en": {
		Language: "en", Path: "/en/", Name: "Arai Hû", Tagline: "Software for stormy weather.",
		Intro:                  "Independent, open tools built to endure difficult work. In Guarani, arai hû means black or dark cloud.",
		ProjectsLabel:          "Built by Arai Hû",
		ProjectsKicker:         "Projects",
		ProjectsIntro:          "One open ecosystem for server-rendered interfaces, API publishing, durable code workflows, and self-hosted monitoring.",
		ProjectAction:          "Explore project",
		OpenSourceLabel:        "Open source",
		NavHome:                "Home",
		NavApps:                "Apps",
		NavLibs:                "Libs",
		NavBlog:                "Blog",
		MenuLabel:              "Menu",
		WorkInProgress:         "WIP",
		FeaturedLabel:          "Featured project",
		FeaturedTitle:          "Interfaces for Go, without leaving the server.",
		FeaturedDescription:    "Goshtoso brings accessible UI components and local assets to server-rendered Go applications.",
		FeaturedAction:         "Explore Goshtoso",
		LibrariesLabel:         "Libraries",
		AppsLabel:              "Applications",
		MoreLabel:              "More stuff",
		MoreIntro:              "Smaller building blocks, focused tools, and shared infrastructure from the same organization.",
		MissionTitle:           "Open software, maintained in the open.",
		MissionCopy:            "Arai Hû builds practical Go software in public. Projects stay independently useful and work better together.",
		MailingTitle:           "Field notes, soon.",
		MailingCopy:            "A mailing list for releases, experiments, and project notes is in progress.",
		MailingPlaceholder:     "you@example.com",
		MailingAction:          "Coming soon",
		ChatTitle:              "Let’s build together.",
		ChatCopy:               "Questions, ideas, or a bug to report? Find the relevant project and start the conversation on GitHub.",
		ChatAction:             "Visit Arai Hû on GitHub",
		Meaning:                "Arai Hû · black or dark cloud in Guarani.",
		SkipLabel:              "Skip to content",
		PrimaryNavigationLabel: "Primary navigation",
		FooterBuiltWith:        "Built with",
		FooterConjunction:      "and",
		HeartChartLabel:        "Three-dimensional parametric heart line.",
		Projects: projects(
			[]string{"Go UI", "API publishing", "Code workflows", "Monitoring"},
			[]string{"A Go UI library for server-rendered applications.", "OpenAPI documentation and publishing workbench.", "Durable workflows for code changes.", "Self-hosted monitoring control plane."},
		),
		MoreProjects: moreProjects(
			[]string{"UI foundations", "Data visualization", "Build tooling"},
			[]string{"Reusable server-rendered application shell patterns built from Goshtoso primitives.", "Chart components for Go and templ applications, from static SVG to interactive exploration.", "Vendor remote build assets with explicit trust and integrity locks."},
		),
	},
	"pt-br": {
		Language: "pt-BR", Path: "/pt-br/", Name: "Arai Hû", Tagline: "Software para passar a trovoada.",
		Intro:                  "Ferramentas independentes e abertas, feitas para resistir ao trabalho difícil. Em guarani, arai hû significa nuvem preta ou escura.",
		ProjectsLabel:          "Criado pela Arai Hû",
		ProjectsKicker:         "Projetos",
		ProjectsIntro:          "Um ecossistema aberto para interfaces renderizadas no servidor, publicação de APIs, workflows duráveis de código e monitoramento auto-hospedado.",
		ProjectAction:          "Explorar projeto",
		OpenSourceLabel:        "Código aberto",
		NavHome:                "Início",
		NavApps:                "Apps",
		NavLibs:                "Bibliotecas",
		NavBlog:                "Blog",
		MenuLabel:              "Menu",
		WorkInProgress:         "Em breve",
		FeaturedLabel:          "Projeto em destaque",
		FeaturedTitle:          "Interfaces para Go, sem sair do servidor.",
		FeaturedDescription:    "Goshtoso leva componentes de UI acessíveis e assets locais a aplicações Go renderizadas no servidor.",
		FeaturedAction:         "Explorar Goshtoso",
		LibrariesLabel:         "Bibliotecas",
		AppsLabel:              "Aplicações",
		MoreLabel:              "Mais projetos",
		MoreIntro:              "Blocos menores, ferramentas focadas e infraestrutura compartilhada pela mesma organização.",
		MissionTitle:           "Software aberto, mantido em público.",
		MissionCopy:            "Arai Hû cria software Go prático de forma aberta. Cada projeto é útil sozinho e funciona melhor em conjunto.",
		MailingTitle:           "Notas de campo, em breve.",
		MailingCopy:            "Uma lista de e-mails para lançamentos, experimentos e notas dos projetos está em desenvolvimento.",
		MailingPlaceholder:     "voce@exemplo.com",
		MailingAction:          "Em breve",
		ChatTitle:              "Vamos construir juntos.",
		ChatCopy:               "Tem uma pergunta, ideia ou bug? Encontre o projeto certo e comece a conversa no GitHub.",
		ChatAction:             "Visitar Arai Hû no GitHub",
		Meaning:                "Arai Hû · nuvem preta ou escura em guarani.",
		SkipLabel:              "Pular para o conteúdo",
		PrimaryNavigationLabel: "Navegação principal",
		FooterBuiltWith:        "Criado com",
		FooterConjunction:      "e",
		HeartChartLabel:        "Linha paramétrica tridimensional em forma de coração.",
		Projects: projects(
			[]string{"UI em Go", "Publicação de APIs", "Workflows de código", "Monitoramento"},
			[]string{"Biblioteca de UI Go para aplicações renderizadas no servidor.", "Documentação OpenAPI e ambiente de publicação.", "Workflows duráveis para mudanças de código.", "Plano de controle de monitoramento auto-hospedado."},
		),
		MoreProjects: moreProjects(
			[]string{"Fundações de UI", "Visualização de dados", "Ferramentas de build"},
			[]string{"Padrões reutilizáveis de shells renderizados no servidor, criados com primitivas Goshtoso.", "Componentes de gráficos para Go e templ, de SVG estático à exploração interativa.", "Distribua assets remotos de build com confiança explícita e locks de integridade."},
		),
	},
	"es": {
		Language: "es", Path: "/es/", Name: "Arai Hû", Tagline: "Software para tiempos de tormenta.",
		Intro:                  "Herramientas independientes y abiertas, creadas para resistir el trabajo difícil. En guaraní, arai hû significa nube negra u oscura.",
		ProjectsLabel:          "Creado por Arai Hû",
		ProjectsKicker:         "Proyectos",
		ProjectsIntro:          "Un ecosistema abierto para interfaces renderizadas en servidor, publicación de APIs, flujos de código durables y monitoreo autoalojado.",
		ProjectAction:          "Explorar proyecto",
		OpenSourceLabel:        "Código abierto",
		NavHome:                "Inicio",
		NavApps:                "Apps",
		NavLibs:                "Bibliotecas",
		NavBlog:                "Blog",
		MenuLabel:              "Menú",
		WorkInProgress:         "En progreso",
		FeaturedLabel:          "Proyecto destacado",
		FeaturedTitle:          "Interfaces para Go, sin salir del servidor.",
		FeaturedDescription:    "Goshtoso lleva componentes de UI accesibles y recursos locales a aplicaciones Go renderizadas en servidor.",
		FeaturedAction:         "Explorar Goshtoso",
		LibrariesLabel:         "Bibliotecas",
		AppsLabel:              "Aplicaciones",
		MoreLabel:              "Más proyectos",
		MoreIntro:              "Bloques más pequeños, herramientas enfocadas e infraestructura compartida por la misma organización.",
		MissionTitle:           "Software abierto, mantenido en público.",
		MissionCopy:            "Arai Hû crea software Go práctico de forma abierta. Cada proyecto es útil por sí solo y funciona mejor en conjunto.",
		MailingTitle:           "Notas de campo, próximamente.",
		MailingCopy:            "Una lista de correo para lanzamientos, experimentos y notas de proyectos está en desarrollo.",
		MailingPlaceholder:     "tu@ejemplo.com",
		MailingAction:          "Próximamente",
		ChatTitle:              "Construyamos juntos.",
		ChatCopy:               "¿Tienes una pregunta, idea o error para reportar? Encuentra el proyecto y empieza la conversación en GitHub.",
		ChatAction:             "Visitar Arai Hû en GitHub",
		Meaning:                "Arai Hû · nube negra u oscura en guaraní.",
		SkipLabel:              "Saltar al contenido",
		PrimaryNavigationLabel: "Navegación principal",
		FooterBuiltWith:        "Creado con",
		FooterConjunction:      "y",
		HeartChartLabel:        "Línea paramétrica tridimensional con forma de corazón.",
		Projects: projects(
			[]string{"UI en Go", "Publicación de APIs", "Flujos de código", "Monitoreo"},
			[]string{"Biblioteca de UI Go para aplicaciones renderizadas en servidor.", "Documentación OpenAPI y espacio de publicación.", "Flujos de trabajo durables para cambios de código.", "Plano de control de monitoreo autoalojado."},
		),
		MoreProjects: moreProjects(
			[]string{"Fundamentos de UI", "Visualización de datos", "Herramientas de build"},
			[]string{"Patrones reutilizables de shells renderizados en servidor, creados con primitivas Goshtoso.", "Componentes de gráficos para Go y templ, desde SVG estático hasta exploración interactiva.", "Distribuye recursos remotos de build con confianza explícita y bloqueos de integridad."},
		),
	},
}

func projects(categories, descriptions []string) []Project {
	return []Project{
		{Name: "Goshtoso", Category: categories[0], Description: descriptions[0], URL: "https://goshtoso.araihu.com", MarkURL: brandAssetBase + "goshtoso-icon-transparent.svg" + brandAssetRevision, Status: "BETA"},
		{Name: "Manja", Category: categories[1], Description: descriptions[1], URL: "https://manja.araihu.com", MarkURL: brandAssetBase + "manja-icon-transparent.svg" + brandAssetRevision, Status: "WIP"},
		{Name: "Pajé", Category: categories[2], Description: descriptions[2], URL: "https://paje.araihu.com", MarkURL: brandAssetBase + "paje-icon-transparent.svg" + brandAssetRevision, Status: "WIP"},
		{Name: "X-9", Category: categories[3], Description: descriptions[3], URL: "https://x9.araihu.com", MarkURL: brandAssetBase + "x9-icon-transparent.svg" + brandAssetRevision, Status: "WIP"},
	}
}

func moreProjects(categories, descriptions []string) []Project {
	return []Project{
		{Name: "Goshtoso App Shells", Category: categories[0], Description: descriptions[0], URL: "https://github.com/araihu/goshtoso-app-shells"},
		{Name: "Goshtoso Charts", Category: categories[1], Description: descriptions[1], URL: "https://github.com/araihu/goshtoso-charts", Status: "ALPHA"},
		{Name: "Muamba", Category: categories[2], Description: descriptions[2], URL: "https://github.com/araihu/muamba"},
	}
}

// Locales returns all generated locales. English is the fallback.
func Locales() []Content { return []Content{contents["en"], contents["pt-br"], contents["es"]} }

// ChartFragmentURL returns the localized static HTMX fragment for a page.
func ChartFragmentURL(content Content) string { return "/fragments" + content.Path + "charts.html" }
