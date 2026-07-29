package site

const CanonicalSiteURL = "https://araihu.com"

// Locale identifies one language version of a public page.
type Locale struct {
	Key, Language, OGLocale, Label string
}

// PageKind identifies a public page family.
type PageKind string

const (
	PageHome    PageKind = "home"
	PageBrand   PageKind = "brand"
	PageLicense PageKind = "license"
)

// Alternate is a reciprocal localized URL for a public page.
type Alternate struct {
	Language, URL string
}

// PageMeta contains page identity independent of a render request.
type PageMeta struct {
	Kind                                   PageKind
	Locale                                 Locale
	Path, CanonicalURL, Title, Description string
	SocialImageURL, Robots                 string
	Alternates                             []Alternate
	StructuredData                         any
}

// LocaleLink is one language switcher destination.
type LocaleLink struct {
	Locale Locale
	URL    string
}

// Navigation contains links shared by a page's navigation controls.
type Navigation struct {
	Locales []LocaleLink
}

// Page is one static, localized public page.
type Page struct {
	Meta       PageMeta
	Navigation Navigation
	Home       *HomeContent
	Brand      *BrandContent
	License    *LicenseContent
}

// BrandContent is localized copy for a brand guidance page.
type BrandContent struct {
	Heading, Lead                         string
	IdentityTitle, IdentityBody           string
	VariantsTitle, VariantsLead           string
	MinimumSizeTitle, MinimumSizeBody     string
	ClearSpaceTitle, ClearSpaceBody       string
	IncorrectUseTitle                     string
	IncorrectUses                         []string
	ContextsTitle, WebTitle, WebBody      string
	MobileTitle, MobileBody               string
	DownloadsTitle, DownloadsLead         string
	Downloads                             []BrandDownload
	IntegrationTitle, IntegrationBody     string
	IntegrationCode                       string
	AttributionTitle, AttributionBody     string
	LicenseTitle, LicenseBody, LicenseCTA string
	Variants                              []BrandVariant
}

// LicenseContent is localized copy for a license page.
type LicenseContent struct {
	Heading, Lead                                       string
	VersionLabel, Version                               string
	EffectiveLabel, EffectiveDate, EffectiveDateDisplay string
	AuthorityNotice, AuthorityLinkLabel                 string
	BrandTermsTitle, HeroiconsTitle, HeroiconsBody      string
	AllowedTitle, PermissionTitle                       string
	Allowed, PermissionRequired                         []string
	NoEndorsementTitle, NoEndorsementBody               string
	NoticesTitle, NoticesBody                           string
	ContactTitle, ContactBody                           string
}

// BrandVariant is one representative approved identity treatment.
type BrandVariant struct {
	Name, Description, AssetURL, SurfaceClass string
}

// BrandDownload is one release-pinned downloadable artifact.
type BrandDownload struct {
	Label, Detail, URL string
}

// ChromeContent localizes shared site navigation and footer copy.
type ChromeContent struct {
	SkipLabel, PrimaryNavigationLabel, LanguageLabel string
	HomeLabel, BrandLabel, LicenseLabel              string
	FooterBuild, Meaning                             string
}
