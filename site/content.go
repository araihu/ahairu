package site

const brandAssetBase = BrandAssetsPublicPrefix + "icons/brand/"

const araihuIconURL = brandAssetBase + "araihu-icon-adaptive-transparent-optical.svg"

// Content is one localized version of the organization site.
type Content struct {
	Language               string
	Path                   string
	Name                   string
	PageTitle              string
	CanonicalURL           string
	OpenGraphLocale        string
	SocialDescription      string
	SocialImageAlt         string
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
	FeaturedFamilyLabel    string
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

// HomeContent is the typed home-page model used by the public page registry.
type HomeContent = Content

// Project is a maintained AraiHu project.
type Project struct {
	Name                string
	Category            string
	ApplicationCategory string
	OperatingSystem     string
	Description         string
	URL                 string
	MarkURL             string
	MarkText            string
	Status              string
}

var contents = map[string]Content{
	"en": {
		Language: "en", Path: "/en/", Name: "Arai Hû", Tagline: "Software for stormy weather.",
		PageTitle:              "Arai Hû — Software for stormy weather.",
		CanonicalURL:           "https://araihu.com/en/",
		OpenGraphLocale:        "en_US",
		SocialDescription:      "Independent, open software built to endure difficult work.",
		SocialImageAlt:         "Arai Hû mark over a dark storm at sea.",
		Intro:                  "Independent, open tools built to endure difficult work. In Guarani, arai hû means black or dark cloud.",
		ProjectsLabel:          "Built by Arai Hû",
		ProjectsKicker:         "Projects",
		ProjectsIntro:          "One open ecosystem for server-rendered interfaces, API publishing, self-hosted monitoring, and service discovery.",
		ProjectAction:          "Explore project",
		OpenSourceLabel:        "Open source",
		NavHome:                "Home",
		NavApps:                "Apps",
		NavLibs:                "Libs",
		NavBlog:                "Blog",
		MenuLabel:              "Menu",
		WorkInProgress:         "WIP",
		FeaturedLabel:          "Featured family",
		FeaturedTitle:          "The server-rendered UI family.",
		FeaturedDescription:    "Goshtoso provides accessible primitives, App Shells composes them into layouts, and Charts brings data visualization to Go and templ.",
		FeaturedAction:         "Explore Goshtoso",
		FeaturedFamilyLabel:    "Goshtoso family",
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
			[]string{"Go UI", "API publishing", "Monitoring", "Homelab dashboard"},
			[]string{"A Go UI library for server-rendered applications.", "OpenAPI documentation and publishing workbench.", "Self-hosted monitoring control plane.", "A self-hosted homepage with Kubernetes, Docker, and Podman service discovery and basic host telemetry."},
		),
		MoreProjects: moreProjects(
			[]string{"UI foundations", "Data visualization", "Build tooling", "Markdown publishing"},
			[]string{"Reusable server-rendered application shell patterns built from Goshtoso primitives.", "Chart components for Go and templ applications, from static SVG to interactive exploration.", "Vendor remote build assets with explicit trust and integrity locks.", "Publish one Markdown source in the format your project needs."},
		),
	},
	"pt-br": {
		Language: "pt-BR", Path: "/pt-br/", Name: "Arai Hû", Tagline: "Software para passar a trovoada.",
		PageTitle:              "Arai Hû — Software para passar a trovoada.",
		CanonicalURL:           "https://araihu.com/pt-br/",
		OpenGraphLocale:        "pt_BR",
		SocialDescription:      "Ferramentas independentes e abertas, feitas para resistir ao trabalho difícil.",
		SocialImageAlt:         "Marca da Arai Hû sobre uma tempestade escura no mar.",
		Intro:                  "Ferramentas independentes e abertas, feitas para resistir ao trabalho difícil. Em guarani, arai hû significa nuvem preta ou escura.",
		ProjectsLabel:          "Criado pela Arai Hû",
		ProjectsKicker:         "Projetos",
		ProjectsIntro:          "Um ecossistema aberto para interfaces renderizadas no servidor, publicação de APIs, monitoramento auto-hospedado e descoberta de serviços.",
		ProjectAction:          "Explorar projeto",
		OpenSourceLabel:        "Código aberto",
		NavHome:                "Início",
		NavApps:                "Apps",
		NavLibs:                "Bibliotecas",
		NavBlog:                "Blog",
		MenuLabel:              "Menu",
		WorkInProgress:         "Em breve",
		FeaturedLabel:          "Família em destaque",
		FeaturedTitle:          "A família de UI renderizada no servidor.",
		FeaturedDescription:    "Goshtoso fornece primitivas acessíveis, App Shells as compõe em layouts e Charts leva visualização de dados a Go e templ.",
		FeaturedAction:         "Explorar Goshtoso",
		FeaturedFamilyLabel:    "Família Goshtoso",
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
			[]string{"UI em Go", "Publicação de APIs", "Monitoramento", "Painel de homelab"},
			[]string{"Biblioteca de UI Go para aplicações renderizadas no servidor.", "Documentação OpenAPI e ambiente de publicação.", "Plano de controle de monitoramento auto-hospedado.", "Página inicial auto-hospedada com descoberta de serviços Kubernetes, Docker e Podman e telemetria básica dos hosts."},
		),
		MoreProjects: moreProjects(
			[]string{"Fundações de UI", "Visualização de dados", "Ferramentas de build", "Publicação em Markdown"},
			[]string{"Padrões reutilizáveis de shells renderizados no servidor, criados com primitivas Goshtoso.", "Componentes de gráficos para Go e templ, de SVG estático à exploração interativa.", "Distribua assets remotos de build com confiança explícita e locks de integridade.", "Publique uma única fonte Markdown no formato que seu projeto precisa."},
		),
	},
	"es": {
		Language: "es", Path: "/es/", Name: "Arai Hû", Tagline: "Software para tiempos de tormenta.",
		PageTitle:              "Arai Hû — Software para tiempos de tormenta.",
		CanonicalURL:           "https://araihu.com/es/",
		OpenGraphLocale:        "es_ES",
		SocialDescription:      "Herramientas independientes y abiertas, creadas para resistir el trabajo difícil.",
		SocialImageAlt:         "Marca de Arai Hû sobre una tormenta oscura en el mar.",
		Intro:                  "Herramientas independientes y abiertas, creadas para resistir el trabajo difícil. En guaraní, arai hû significa nube negra u oscura.",
		ProjectsLabel:          "Creado por Arai Hû",
		ProjectsKicker:         "Proyectos",
		ProjectsIntro:          "Un ecosistema abierto para interfaces renderizadas en servidor, publicación de APIs, monitoreo autoalojado y descubrimiento de servicios.",
		ProjectAction:          "Explorar proyecto",
		OpenSourceLabel:        "Código abierto",
		NavHome:                "Inicio",
		NavApps:                "Apps",
		NavLibs:                "Bibliotecas",
		NavBlog:                "Blog",
		MenuLabel:              "Menú",
		WorkInProgress:         "En progreso",
		FeaturedLabel:          "Familia destacada",
		FeaturedTitle:          "La familia de UI renderizada en servidor.",
		FeaturedDescription:    "Goshtoso aporta primitivas accesibles, App Shells las compone en layouts y Charts lleva visualización de datos a Go y templ.",
		FeaturedAction:         "Explorar Goshtoso",
		FeaturedFamilyLabel:    "Familia Goshtoso",
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
			[]string{"UI en Go", "Publicación de APIs", "Monitoreo", "Panel de homelab"},
			[]string{"Biblioteca de UI Go para aplicaciones renderizadas en servidor.", "Documentación OpenAPI y espacio de publicación.", "Plano de control de monitoreo autoalojado.", "Página de inicio autoalojada con descubrimiento de servicios Kubernetes, Docker y Podman y telemetría básica de los hosts."},
		),
		MoreProjects: moreProjects(
			[]string{"Fundamentos de UI", "Visualización de datos", "Herramientas de build", "Publicación en Markdown"},
			[]string{"Patrones reutilizables de shells renderizados en servidor, creados con primitivas Goshtoso.", "Componentes de gráficos para Go y templ, desde SVG estático hasta exploración interactiva.", "Distribuye recursos remotos de build con confianza explícita y bloqueos de integridad.", "Publica una única fuente Markdown en el formato que necesita tu proyecto."},
		),
	},
}

func projects(categories, descriptions []string) []Project {
	return []Project{
		{Name: "Goshtoso", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux", Category: categories[0], Description: descriptions[0], URL: "https://goshtoso.araihu.com", MarkURL: brandAssetBase + "goshtoso-icon-dark-plate-optical.svg", Status: "BETA"},
		{Name: "Manja", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux", Category: categories[1], Description: descriptions[1], URL: "https://manja.araihu.com", MarkURL: brandAssetBase + "manja-icon-dark-plate-optical.svg", Status: "BETA"},
		{Name: "X-9", ApplicationCategory: "UtilitiesApplication", OperatingSystem: "Linux", Category: categories[2], Description: descriptions[2], URL: "https://x9.araihu.com", MarkURL: brandAssetBase + "x9-icon-dark-plate-optical.svg", Status: "ALPHA"},
		{Name: "Balemoh", ApplicationCategory: "UtilitiesApplication", OperatingSystem: "Linux", Category: categories[3], Description: descriptions[3], URL: "https://balemoh.araihu.com", MarkURL: brandAssetBase + "balemoh-icon-dark-plate-optical.svg", Status: "ALPHA"},
	}
}

func moreProjects(categories, descriptions []string) []Project {
	return []Project{
		{Name: "Goshtoso App Shells", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux", Category: categories[0], Description: descriptions[0], URL: "https://github.com/araihu/goshtoso-app-shells", Status: "BETA"},
		{Name: "Goshtoso Charts", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux", Category: categories[1], Description: descriptions[1], URL: "https://charts.goshtoso.araihu.com", Status: "ALPHA"},
		{Name: "Muamba", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux, macOS, Windows", Category: categories[2], Description: descriptions[2], URL: "https://muamba.araihu.com", Status: "ALPHA"},
		{Name: "Margo", ApplicationCategory: "DeveloperApplication", OperatingSystem: "Linux, macOS, Windows", Category: categories[3], Description: descriptions[3], URL: "https://margo.araihu.com", MarkURL: "/assets/visuals/margo-icon.svg", Status: "ALPHA"},
	}
}

// Locales returns all generated locales. English is the fallback.
func homeContent(localeKey string) HomeContent { return contents[localeKey] }

func Locales() []HomeContent { return []HomeContent{contents["en"], contents["pt-br"], contents["es"]} }

// ChartFragmentURL returns the localized static HTMX fragment for a page.
func ChartFragmentURL(content Content) string { return "/fragments" + content.Path + "charts.html" }
