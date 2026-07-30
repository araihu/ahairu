package site

var chromeContent = map[string]ChromeContent{
	"en": {
		SkipLabel: "Skip to content", PrimaryNavigationLabel: "Primary navigation", LanguageLabel: "Language",
		HomeLabel: "Home", BrandLabel: "Brand", LicenseLabel: "License",
		FooterBuild: "Built with Go and Goshtoso.", Meaning: "Arai Hû · black or dark cloud in Guarani.",
	},
	"pt-br": {
		SkipLabel: "Pular para o conteúdo", PrimaryNavigationLabel: "Navegação principal", LanguageLabel: "Idioma",
		HomeLabel: "Início", BrandLabel: "Marca", LicenseLabel: "Licença",
		FooterBuild: "Criado com Go e Goshtoso.", Meaning: "Arai Hû · nuvem preta ou escura em guarani.",
	},
	"es": {
		SkipLabel: "Saltar al contenido", PrimaryNavigationLabel: "Navegación principal", LanguageLabel: "Idioma",
		HomeLabel: "Inicio", BrandLabel: "Marca", LicenseLabel: "Licencia",
		FooterBuild: "Creado con Go y Goshtoso.", Meaning: "Arai Hû · nube negra u oscura en guaraní.",
	},
}

func sharedContent(localeKey string) ChromeContent { return chromeContent[localeKey] }

func brandContent(localeKey string) BrandContent {
	content := brandContents[localeKey]
	content.Variants = localizedBrandVariants(localeKey)
	content.Downloads = localizedBrandDownloads(localeKey)
	return content
}

var brandContents = map[string]BrandContent{
	"en": {
		Heading: "Brand guidance", Lead: "A practical identity system for clear, consistent Arai Hû communication across websites and applications.",
		IdentityTitle: "Identity", IdentityBody: "Arai Hû means black or dark cloud in Guarani. The cloud and lime signal form the primary mark; the outlined wordmark is the preferred signature when space allows.",
		VariantsTitle: "Approved variants", VariantsLead: "Choose by surface, not preference. Keep embedded colors unchanged except for the designated monochrome asset.",
		MinimumSizeTitle: "Minimum size", MinimumSizeBody: "Use the optical icon at 16 px or larger on the web and 24 dp or larger in interfaces. Use the logo at 120 px wide or larger; below that, switch to the icon.",
		ClearSpaceTitle: "Clear space", ClearSpaceBody: "Keep one lightning-bolt width of clear space around every side of the mark. No text, controls, borders, or crop edges may enter this area.",
		IncorrectUseTitle: "Incorrect use", IncorrectUses: []string{"Do not stretch, rotate, redraw, crop, or rearrange the mark.", "Do not recolor protected variants or add gradients, shadows, outlines, or effects.", "Do not place the mark on low-contrast or visually noisy surfaces.", "Do not use an Arai Hû mark as another product or organization identity."},
		ContextsTitle: "Web and mobile examples", WebTitle: "Web", WebBody: "Use the transparent optical icon in navigation and the outlined logo in wider headers. Use release web icons for favicons, touch icons, and manifests.",
		MobileTitle: "Mobile", MobileBody: "Use launcher-framed assets for app stores and launchers. Preserve the supplied safe zone; do not crop an optical icon into a launcher shape.",
		DownloadsTitle: "Download", DownloadsLead: "Files below are immutable v0.1.1 release artifacts. Keep NOTICE and the applicable license with redistributed files.",
		IntegrationTitle: "Sprite and Goshtoso", IntegrationBody: "The brand sprite exposes canonical symbol names. Goshtoso renders a safe external SVG reference; catalog.json remains the source for generated client bindings.",
		IntegrationCode:  `icon.Icon(icon.Config{SpriteURL: brandicons.SpriteURL, Symbol: brandicons.IconAraihuIconMonochromeTransparentOptical, Label: "Arai Hû"})`,
		AttributionTitle: "Attribution", AttributionBody: "Identify the asset as an Arai Hû mark, retain NOTICE, and keep third-party license notices. Attribution does not grant endorsement.",
		LicenseTitle: "Use the identity responsibly", LicenseBody: "The brand terms explain permitted unmodified use and when written permission is required.", LicenseCTA: "Read license and distribution terms",
	},
	"pt-br": {
		Heading: "Guia de marca", Lead: "Um sistema de identidade prático para comunicar Arai Hû com clareza e consistência em sites e aplicativos.",
		IdentityTitle: "Identidade", IdentityBody: "Arai Hû significa nuvem preta ou escura em guarani. A nuvem e o sinal lima formam a marca principal; o logotipo delineado é a assinatura preferida quando houver espaço.",
		VariantsTitle: "Variantes aprovadas", VariantsLead: "Escolha conforme a superfície. Preserve as cores incorporadas, exceto no arquivo monocromático designado.",
		MinimumSizeTitle: "Tamanho mínimo", MinimumSizeBody: "Use o ícone óptico com pelo menos 16 px na web e 24 dp em interfaces. Use o logotipo com pelo menos 120 px de largura; abaixo disso, use o ícone.",
		ClearSpaceTitle: "Área de proteção", ClearSpaceBody: "Mantenha o equivalente à largura do raio livre em todos os lados. Texto, controles, bordas e cortes não podem entrar nessa área.",
		IncorrectUseTitle: "Usos incorretos", IncorrectUses: []string{"Não estique, gire, redesenhe, corte ou reorganize a marca.", "Não altere a cor das variantes protegidas nem adicione gradientes, sombras ou efeitos.", "Não aplique a marca sobre superfícies de baixo contraste ou visualmente ruidosas.", "Não use uma marca Arai Hû como identidade de outro produto ou organização."},
		ContextsTitle: "Exemplos para web e mobile", WebTitle: "Web", WebBody: "Use o ícone óptico transparente na navegação e o logotipo em cabeçalhos amplos. Use os arquivos web da versão para favicon, touch icon e manifesto.",
		MobileTitle: "Mobile", MobileBody: "Use os arquivos com moldura de launcher em lojas e launchers. Preserve a área segura fornecida; não corte um ícone óptico para criar um launcher.",
		DownloadsTitle: "Baixar", DownloadsLead: "Os arquivos abaixo são artefatos imutáveis da versão v0.1.1. Inclua NOTICE e a licença aplicável ao redistribuí-los.",
		IntegrationTitle: "Sprite e Goshtoso", IntegrationBody: "O sprite de marca expõe nomes canônicos. Goshtoso renderiza uma referência SVG externa segura; catalog.json orienta a geração de bindings nos clientes.",
		IntegrationCode:  `icon.Icon(icon.Config{SpriteURL: brandicons.SpriteURL, Symbol: brandicons.IconAraihuIconMonochromeTransparentOptical, Label: "Arai Hû"})`,
		AttributionTitle: "Atribuição", AttributionBody: "Identifique o arquivo como marca Arai Hû, preserve NOTICE e os avisos de licenças de terceiros. A atribuição não concede endosso.",
		LicenseTitle: "Use a identidade com responsabilidade", LicenseBody: "Os termos de marca explicam o uso não modificado permitido e quando é necessária autorização por escrito.", LicenseCTA: "Ler licença e termos de distribuição",
	},
	"es": {
		Heading: "Guía de marca", Lead: "Un sistema de identidad práctico para comunicar Arai Hû con claridad y consistencia en sitios y aplicaciones.",
		IdentityTitle: "Identidad", IdentityBody: "Arai Hû significa nube negra u oscura en guaraní. La nube y la señal lima forman la marca principal; el logotipo delineado es la firma preferida cuando hay espacio.",
		VariantsTitle: "Variantes aprobadas", VariantsLead: "Elige según la superficie. Conserva los colores incorporados, excepto en el archivo monocromático designado.",
		MinimumSizeTitle: "Tamaño mínimo", MinimumSizeBody: "Usa el icono óptico a partir de 16 px en web y 24 dp en interfaces. Usa el logotipo a partir de 120 px de ancho; por debajo, cambia al icono.",
		ClearSpaceTitle: "Área de protección", ClearSpaceBody: "Mantén un ancho de rayo libre en cada lado. Texto, controles, bordes y recortes no pueden entrar en esta área.",
		IncorrectUseTitle: "Usos incorrectos", IncorrectUses: []string{"No estires, gires, redibujes, recortes ni reorganices la marca.", "No cambies el color de variantes protegidas ni añadas degradados, sombras o efectos.", "No coloques la marca sobre superficies de bajo contraste o con ruido visual.", "No uses una marca Arai Hû como identidad de otro producto u organización."},
		ContextsTitle: "Ejemplos para web y móvil", WebTitle: "Web", WebBody: "Usa el icono óptico transparente en navegación y el logotipo en cabeceras amplias. Usa los archivos web de la versión para favicon, icono táctil y manifiesto.",
		MobileTitle: "Móvil", MobileBody: "Usa archivos con marco de launcher en tiendas y launchers. Conserva la zona segura suministrada; no recortes un icono óptico para crear un launcher.",
		DownloadsTitle: "Descargar", DownloadsLead: "Los archivos siguientes son artefactos inmutables de v0.1.1. Incluye NOTICE y la licencia aplicable al redistribuirlos.",
		IntegrationTitle: "Sprite y Goshtoso", IntegrationBody: "El sprite de marca expone nombres canónicos. Goshtoso renderiza una referencia SVG externa segura; catalog.json guía la generación de bindings del cliente.",
		IntegrationCode:  `icon.Icon(icon.Config{SpriteURL: brandicons.SpriteURL, Symbol: brandicons.IconAraihuIconMonochromeTransparentOptical, Label: "Arai Hû"})`,
		AttributionTitle: "Atribución", AttributionBody: "Identifica el archivo como marca Arai Hû, conserva NOTICE y los avisos de licencias de terceros. La atribución no concede respaldo.",
		LicenseTitle: "Usa la identidad con responsabilidad", LicenseBody: "Los términos de marca explican el uso no modificado permitido y cuándo se necesita autorización escrita.", LicenseCTA: "Leer licencia y términos de distribución",
	},
}

func localizedBrandVariants(localeKey string) []BrandVariant {
	names := map[string][][2]string{
		"en":    {{"Light", "For paper and light surfaces."}, {"Dark", "For midnight and dark surfaces."}, {"Monochrome", "The only color-inheriting variant."}, {"Tinted", "Designed cobalt and lime treatment."}},
		"pt-br": {{"Clara", "Para papel e superfícies claras."}, {"Escura", "Para superfícies noturnas e escuras."}, {"Monocromática", "Única variante que herda cor."}, {"Tingida", "Tratamento projetado em cobalto e lima."}},
		"es":    {{"Clara", "Para papel y superficies claras."}, {"Oscura", "Para superficies nocturnas y oscuras."}, {"Monocromática", "Única variante que hereda color."}, {"Tintada", "Tratamiento diseñado en cobalto y lima."}},
	}[localeKey]
	paths := []string{"light-transparent-optical.svg", "dark-transparent-optical.svg", "monochrome-transparent-optical.svg", "tinted-transparent-optical.svg"}
	surfaces := []string{"variant-light", "variant-dark", "variant-signal", "variant-tinted"}
	variants := make([]BrandVariant, len(paths))
	for index := range paths {
		variants[index] = BrandVariant{Name: names[index][0], Description: names[index][1], AssetURL: BrandAssetsPublicPrefix + "brand/araihu/logo/" + paths[index], SurfaceClass: surfaces[index]}
	}
	return variants
}

func localizedBrandDownloads(localeKey string) []BrandDownload {
	labels := map[string][]string{
		"en":    {"Light logo", "Dark logo", "Monochrome logo", "Tinted logo", "Brand sprite", "catalog.json", "checksums.txt", "NOTICE"},
		"pt-br": {"Logotipo claro", "Logotipo escuro", "Logotipo monocromático", "Logotipo tingido", "Sprite de marca", "catalog.json", "checksums.txt", "NOTICE"},
		"es":    {"Logotipo claro", "Logotipo oscuro", "Logotipo monocromático", "Logotipo tintado", "Sprite de marca", "catalog.json", "checksums.txt", "NOTICE"},
	}[localeKey]
	paths := []string{"brand/araihu/logo/light-transparent-optical.svg", "brand/araihu/logo/dark-transparent-optical.svg", "brand/araihu/logo/monochrome-transparent-optical.svg", "brand/araihu/logo/tinted-transparent-optical.svg", "icons/brand/sprite.svg", "catalog.json", "checksums.txt", "NOTICE"}
	details := []string{"SVG · protected color", "SVG · protected color", "SVG · currentColor", "SVG · protected color", "SVG symbols", "JSON · schema v1", "SHA-256", "License notices"}
	downloads := make([]BrandDownload, len(paths))
	for index := range paths {
		downloads[index] = BrandDownload{Label: labels[index], Detail: details[index], URL: BrandAssetsPublicPrefix + paths[index]}
	}
	return downloads
}

func licenseContent(localeKey string) LicenseContent { return licenseContents[localeKey] }

var licenseContents = map[string]LicenseContent{
	"en": {
		Heading: "License", Lead: "Distribution rules for Arai Hû identity assets and bundled third-party icons.",
		VersionLabel: "Version", Version: licenseVersion, EffectiveLabel: "Effective", EffectiveDate: licenseEffectiveDate, EffectiveDateDisplay: "29 July 2026",
		BrandTermsTitle: "Arai Hû brand terms", HeroiconsTitle: "Heroicons MIT license", HeroiconsBody: "Heroicons UI artwork remains available under the MIT license included with the asset release. These brand terms do not replace that license.",
		AllowedTitle: "Allowed without written permission", Allowed: []string{"Unmodified integration in an Arai Hû website or application.", "Unmodified use in documentation that accurately refers to Arai Hû or its projects.", "Redistribution with a permitted integration or documentation package when NOTICE and applicable license notices remain included.", "Reasonable resizing and file-format conversion that preserve appearance, proportions, clear space, and meaning."},
		PermissionTitle: "Written permission required", PermissionRequired: []string{"Modified marks or altered protected colors.", "Standalone redistribution as a logo pack, icon pack, template, or competing asset library.", "Merchandise, physical goods, paid promotional material, or commercial sponsorship.", "Use as another identity for a product, service, organization, community, or account.", "Any presentation suggesting implied affiliation, certification, sponsorship, or approval."},
		NoEndorsementTitle: "No endorsement", NoEndorsementBody: "Permission to display or redistribute an asset does not imply affiliation, sponsorship, certification, or endorsement by Arai Hû.",
		NoticesTitle: "Notices and attribution", NoticesBody: "Keep the release NOTICE and all applicable third-party notices with redistributed assets. Identify Arai Hû marks accurately and do not remove ownership or source information.",
		ContactTitle: "Request permission", ContactBody: "For modified marks, standalone distribution, merchandise, or uses not covered here, request written permission before publication or production.",
	},
	"pt-br": {
		Heading: "Licença", Lead: "Regras de distribuição dos ativos de identidade Arai Hû e ícones de terceiros incluídos.",
		VersionLabel: "Versão", Version: licenseVersion, EffectiveLabel: "Vigente em", EffectiveDate: licenseEffectiveDate, EffectiveDateDisplay: "29 de julho de 2026",
		AuthorityNotice: "A versão em inglês rege estes termos. Esta tradução é apenas informativa.", AuthorityLinkLabel: "Consultar a versão em inglês",
		BrandTermsTitle: "Termos de marca Arai Hû", HeroiconsTitle: "Licença MIT do Heroicons", HeroiconsBody: "Os ícones de UI Heroicons permanecem disponíveis sob a licença MIT incluída na versão dos ativos. Estes termos de marca não substituem essa licença.",
		AllowedTitle: "Permitido sem autorização escrita", Allowed: []string{"Integração não modificada em site ou aplicativo Arai Hû.", "Uso não modificado em documentação que se refira corretamente à Arai Hû ou a seus projetos.", "Redistribuição junto de integração ou documentação permitida, mantendo NOTICE e os avisos de licença aplicáveis.", "Redimensionamento razoável e conversão de formato que preservem aparência, proporções, área de proteção e significado."},
		PermissionTitle: "Autorização escrita necessária", PermissionRequired: []string{"Marcas modificadas ou cores protegidas alteradas.", "Redistribuição independente como pacote de logos, ícones, template ou biblioteca de ativos concorrente.", "Produtos físicos, mercadorias, material promocional pago ou patrocínio comercial.", "Uso como outra identidade de produto, serviço, organização, comunidade ou conta.", "Apresentação que sugira afiliação, certificação, patrocínio ou aprovação implícita."},
		NoEndorsementTitle: "Sem endosso", NoEndorsementBody: "A permissão para exibir ou redistribuir um ativo não implica afiliação, patrocínio, certificação ou endosso pela Arai Hû.",
		NoticesTitle: "Avisos e atribuição", NoticesBody: "Mantenha o NOTICE da versão e todos os avisos de terceiros aplicáveis junto aos ativos redistribuídos. Identifique corretamente as marcas Arai Hû.",
		ContactTitle: "Solicitar autorização", ContactBody: "Para marcas modificadas, distribuição independente, mercadorias ou usos não cobertos, solicite autorização escrita antes da publicação ou produção.",
	},
	"es": {
		Heading: "Licencia", Lead: "Reglas de distribución para los activos de identidad Arai Hû y los iconos de terceros incluidos.",
		VersionLabel: "Versión", Version: licenseVersion, EffectiveLabel: "Vigente desde", EffectiveDate: licenseEffectiveDate, EffectiveDateDisplay: "29 de julio de 2026",
		AuthorityNotice: "La versión en inglés rige estos términos. Esta traducción es únicamente informativa.", AuthorityLinkLabel: "Consultar la versión en inglés",
		BrandTermsTitle: "Términos de marca Arai Hû", HeroiconsTitle: "Licencia MIT de Heroicons", HeroiconsBody: "Los iconos de UI Heroicons siguen disponibles bajo la licencia MIT incluida con la versión de activos. Estos términos de marca no sustituyen esa licencia.",
		AllowedTitle: "Permitido sin autorización escrita", Allowed: []string{"Integración sin modificaciones en un sitio o aplicación Arai Hû.", "Uso sin modificaciones en documentación que se refiera correctamente a Arai Hû o sus proyectos.", "Redistribución junto con una integración o documentación permitida, conservando NOTICE y los avisos de licencia aplicables.", "Cambio razonable de tamaño y formato que conserve apariencia, proporciones, área de protección y significado."},
		PermissionTitle: "Se requiere autorización escrita", PermissionRequired: []string{"Marcas modificadas o cambios en colores protegidos.", "Redistribución independiente como paquete de logos, iconos, plantilla o biblioteca de activos competidora.", "Mercancía, productos físicos, material promocional pagado o patrocinio comercial.", "Uso como otra identidad de producto, servicio, organización, comunidad o cuenta.", "Presentación que sugiera afiliación, certificación, patrocinio o aprobación implícita."},
		NoEndorsementTitle: "Sin respaldo", NoEndorsementBody: "El permiso para mostrar o redistribuir un activo no implica afiliación, patrocinio, certificación ni respaldo de Arai Hû.",
		NoticesTitle: "Avisos y atribución", NoticesBody: "Conserva el NOTICE de la versión y todos los avisos de terceros aplicables con los activos redistribuidos. Identifica correctamente las marcas Arai Hû.",
		ContactTitle: "Solicitar autorización", ContactBody: "Para marcas modificadas, distribución independiente, mercancía o usos no cubiertos, solicita autorización escrita antes de publicar o producir.",
	},
}
