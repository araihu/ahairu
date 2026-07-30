// Package assetbundle validates and assembles immutable Arai Hu asset releases.
package assetbundle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	releaseTag = regexp.MustCompile(`^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-(?:(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
	sha256Hex  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	lowerKebab = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
)

// Bundle is a validated immutable release and channel bundle.
type Bundle struct {
	Releases []string
	Latest   Channel
	Default  Channel
	Current  Channel
}

// Channel is the release index extracted from a full channel document.
type Channel struct {
	SchemaVersion int    `json:"schemaVersion"`
	Release       string `json:"release"`
	Digest        string `json:"digest"`
}

type channelDocument struct {
	SchemaVersion  int               `json:"schemaVersion"`
	RuntimeVersion int               `json:"runtimeVersion"`
	Release        string            `json:"release"`
	Source         string            `json:"source"`
	Theme          resolvedTheme     `json:"theme"`
	Campaign       *resolvedCampaign `json:"campaign,omitempty"`
	Digest         string            `json:"digest"`
}

type resolvedTheme struct {
	ID     string `json:"id"`
	CSSURL string `json:"cssUrl"`
}

type resolvedCampaign struct {
	ID     string         `json:"id"`
	Toggle resolvedToggle `json:"toggle"`
	Brand  resolvedBrand  `json:"brand"`
}

type resolvedToggle struct {
	EnabledIcon  resolvedIcon `json:"enabledIcon"`
	DisabledIcon resolvedIcon `json:"disabledIcon"`
}

type resolvedBrand struct {
	Logo resolvedAsset `json:"logo"`
	Icon resolvedAsset `json:"icon"`
}

type resolvedAsset struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type resolvedIcon struct {
	ID           string `json:"id"`
	Mode         string `json:"mode"`
	URL          string `json:"url"`
	SpriteSymbol string `json:"spriteSymbol,omitempty"`
}

type releaseDocument struct {
	SchemaVersion    int           `json:"schemaVersion"`
	Release          string        `json:"release"`
	IdentityRevision int           `json:"identityRevision"`
	RuntimeVersion   int           `json:"runtimeVersion"`
	CatalogSHA256    string        `json:"catalogSha256"`
	ThemesSHA256     string        `json:"themesSha256"`
	CampaignsSHA256  string        `json:"campaignsSha256"`
	Files            []releaseFile `json:"files"`
}

type releaseFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Validate verifies a complete public bundle before it can be assembled.
func Validate(source fs.FS) (Bundle, error) {
	if source == nil {
		return Bundle{}, errors.New("asset bundle source is nil")
	}
	files := make(map[string][]byte)
	releases := make(map[string]struct{})
	if err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		if !safePath(name) {
			return fmt.Errorf("asset bundle path %q is invalid", name)
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("asset bundle path %q is a symbolic link", name)
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect asset bundle path %q: %w", name, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("asset bundle path %q is a symbolic link", name)
		}
		if !knownPath(name, info.IsDir(), releases) {
			return fmt.Errorf("asset bundle path %q is outside known layout", name)
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("asset bundle path %q is not a regular file", name)
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return fmt.Errorf("read asset bundle path %q: %w", name, err)
		}
		files[name] = data
		return nil
	}); err != nil {
		return Bundle{}, err
	}

	if _, ok := files["campaign/v1.js"]; !ok {
		return Bundle{}, errors.New("asset bundle misses campaign/v1.js")
	}
	for _, channel := range []string{"latest", "default", "current"} {
		if _, ok := files["releases/"+channel+".json"]; !ok {
			return Bundle{}, fmt.Errorf("asset bundle misses releases/%s.json", channel)
		}
	}
	if len(releases) == 0 {
		return Bundle{}, errors.New("asset bundle has no immutable releases")
	}

	ordered := make([]string, 0, len(releases))
	for release := range releases {
		if err := validateRelease(release, files); err != nil {
			return Bundle{}, err
		}
		ordered = append(ordered, release)
	}
	sort.Strings(ordered)

	channels := make(map[string]Channel, 3)
	for _, name := range []string{"latest", "default", "current"} {
		channel, err := validateChannel(name, files["releases/"+name+".json"], releases, files)
		if err != nil {
			return Bundle{}, err
		}
		channels[name] = channel
	}
	return Bundle{Releases: ordered, Latest: channels["latest"], Default: channels["default"], Current: channels["current"]}, nil
}

func knownPath(name string, directory bool, releases map[string]struct{}) bool {
	parts := strings.Split(name, "/")
	switch parts[0] {
	case "campaign":
		return (directory && name == "campaign") || (!directory && name == "campaign/v1.js")
	case "releases":
		if name == "releases" {
			return directory
		}
		if len(parts) == 2 {
			if !directory {
				return parts[1] == "latest.json" || parts[1] == "default.json" || parts[1] == "current.json"
			}
			if !releaseTag.MatchString(parts[1]) {
				return false
			}
			if _, duplicate := releases[parts[1]]; duplicate {
				return false
			}
			releases[parts[1]] = struct{}{}
			return true
		}
		if len(parts) < 3 || !releaseTag.MatchString(parts[1]) {
			return false
		}
		_, found := releases[parts[1]]
		return found
	default:
		return false
	}
}

func validateRelease(release string, files map[string][]byte) error {
	prefix := "releases/" + release + "/"
	releaseJSON, ok := files[prefix+"release.json"]
	if !ok {
		return fmt.Errorf("immutable release %q misses release.json", release)
	}
	checksums, ok := files[prefix+"checksums.txt"]
	if !ok {
		return fmt.Errorf("immutable release %q misses checksums.txt", release)
	}
	var document releaseDocument
	if err := decodeStrict(releaseJSON, &document); err != nil {
		return fmt.Errorf("decode immutable release %q release.json: %w", release, err)
	}
	if err := validateReleaseDocument(release, document, files); err != nil {
		return err
	}
	want, err := parseChecksums(checksums)
	if err != nil {
		return fmt.Errorf("decode immutable release %q checksums.txt: %w", release, err)
	}
	actual := make(map[string][]byte)
	for name, data := range files {
		if strings.HasPrefix(name, prefix) && name != prefix+"checksums.txt" {
			actual[strings.TrimPrefix(name, prefix)] = data
		}
	}
	if len(want) != len(actual) {
		return fmt.Errorf("immutable release %q checksum entries = %d, want %d", release, len(want), len(actual))
	}
	for name, data := range actual {
		digest, found := want[name]
		if !found {
			return fmt.Errorf("immutable release %q checksum misses %q", release, name)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != digest {
			return fmt.Errorf("immutable release %q checksum mismatch for %q", release, name)
		}
	}
	return nil
}

func validateReleaseDocument(release string, document releaseDocument, files map[string][]byte) error {
	if document.SchemaVersion != 1 || document.Release != release || document.IdentityRevision < 1 || document.RuntimeVersion != 1 {
		return fmt.Errorf("immutable release %q has invalid release metadata", release)
	}
	for _, item := range []struct{ name, value string }{{"catalogSha256", document.CatalogSHA256}, {"themesSha256", document.ThemesSHA256}, {"campaignsSha256", document.CampaignsSHA256}} {
		if !sha256Hex.MatchString(item.value) {
			return fmt.Errorf("immutable release %q has invalid %s", release, item.name)
		}
	}
	if len(document.Files) == 0 {
		return fmt.Errorf("immutable release %q has empty inventory", release)
	}
	previous := ""
	inventory := make(map[string]releaseFile, len(document.Files))
	for _, item := range document.Files {
		if !safePath(item.Path) || item.Path == "release.json" || item.Path == "checksums.txt" || !sha256Hex.MatchString(item.SHA256) || item.Size < 0 || (previous != "" && compareReleasePaths(previous, item.Path) >= 0) {
			return fmt.Errorf("immutable release %q has invalid inventory entry %q", release, item.Path)
		}
		previous = item.Path
		inventory[item.Path] = item
		data, found := files["releases/"+release+"/"+item.Path]
		if !found {
			return fmt.Errorf("immutable release %q inventory misses file %q", release, item.Path)
		}
		sum := sha256.Sum256(data)
		if item.SHA256 != hex.EncodeToString(sum[:]) || item.Size != int64(len(data)) {
			return fmt.Errorf("immutable release %q inventory checksum mismatch for %q", release, item.Path)
		}
	}
	for _, required := range []struct{ name, digest string }{{"catalog.json", document.CatalogSHA256}, {"themes.json", document.ThemesSHA256}, {"campaigns.json", document.CampaignsSHA256}} {
		item, found := inventory[required.name]
		if !found || item.SHA256 != required.digest {
			return fmt.Errorf("immutable release %q metadata checksum mismatch for %q", release, required.name)
		}
	}
	actual := make(map[string]struct{})
	prefix := "releases/" + release + "/"
	for name := range files {
		if strings.HasPrefix(name, prefix) {
			relative := strings.TrimPrefix(name, prefix)
			if relative != "release.json" && relative != "checksums.txt" {
				actual[relative] = struct{}{}
			}
		}
	}
	if len(actual) != len(inventory) {
		return fmt.Errorf("immutable release %q inventory file count = %d, want %d", release, len(inventory), len(actual))
	}
	for name := range actual {
		if _, found := inventory[name]; !found {
			return fmt.Errorf("immutable release %q inventory omits file %q", release, name)
		}
	}
	return nil
}

func compareReleasePaths(left, right string) int {
	order := map[string]int{"catalog.json": 0, "themes.json": 1, "campaigns.json": 2}
	leftOrder, leftKnown := order[left]
	rightOrder, rightKnown := order[right]
	if leftKnown && rightKnown {
		return leftOrder - rightOrder
	}
	if leftKnown {
		return -1
	}
	if rightKnown {
		return 1
	}
	return strings.Compare(left, right)
}

func parseChecksums(data []byte) (map[string]string, error) {
	entries := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || !sha256Hex.MatchString(parts[0]) || !safePath(parts[1]) {
			return nil, fmt.Errorf("invalid checksum line %q", line)
		}
		if _, duplicate := entries[parts[1]]; duplicate {
			return nil, fmt.Errorf("duplicate checksum path %q", parts[1])
		}
		entries[parts[1]] = parts[0]
	}
	return entries, nil
}

func validateChannel(name string, data []byte, releases map[string]struct{}, files map[string][]byte) (Channel, error) {
	var document channelDocument
	if err := decodeStrict(data, &document); err != nil {
		return Channel{}, fmt.Errorf("decode releases/%s.json: %w", name, err)
	}
	if document.SchemaVersion != 1 || document.RuntimeVersion != 1 || !releaseTag.MatchString(document.Release) || !sha256Hex.MatchString(document.Digest) {
		return Channel{}, fmt.Errorf("releases/%s.json has invalid channel metadata", name)
	}
	if _, found := releases[document.Release]; !found {
		return Channel{}, fmt.Errorf("releases/%s.json points outside immutable release %q", name, document.Release)
	}
	if document.Source != "default" && document.Source != "campaign" {
		return Channel{}, fmt.Errorf("releases/%s.json has invalid source", name)
	}
	if (document.Source == "campaign") != (document.Campaign != nil) {
		return Channel{}, fmt.Errorf("releases/%s.json has inconsistent campaign", name)
	}
	if !lowerKebab.MatchString(document.Theme.ID) {
		return Channel{}, fmt.Errorf("releases/%s.json has invalid theme", name)
	}
	if err := validateReleaseURL(document.Release, document.Theme.CSSURL, files); err != nil {
		return Channel{}, fmt.Errorf("releases/%s.json theme is outside immutable release: %w", name, err)
	}
	if document.Campaign != nil {
		if err := validateCampaign(document.Release, *document.Campaign, files); err != nil {
			return Channel{}, fmt.Errorf("releases/%s.json campaign is outside immutable release: %w", name, err)
		}
	}
	canonical := document
	canonical.Digest = ""
	payload, err := encodeCanonical(canonical)
	if err != nil {
		return Channel{}, err
	}
	sum := sha256.Sum256(payload)
	if document.Digest != hex.EncodeToString(sum[:]) {
		return Channel{}, fmt.Errorf("releases/%s.json digest mismatch", name)
	}
	return Channel{SchemaVersion: document.SchemaVersion, Release: document.Release, Digest: document.Digest}, nil
}

func validateCampaign(release string, campaign resolvedCampaign, files map[string][]byte) error {
	if !lowerKebab.MatchString(campaign.ID) {
		return errors.New("invalid campaign id")
	}
	for _, asset := range []string{campaign.Toggle.EnabledIcon.URL, campaign.Toggle.DisabledIcon.URL, campaign.Brand.Logo.URL, campaign.Brand.Icon.URL} {
		if err := validateReleaseURL(release, asset, files); err != nil {
			return err
		}
	}
	return nil
}

func validateReleaseURL(release, value string, files map[string][]byte) error {
	prefix := "/assets/releases/" + release + "/"
	if !strings.HasPrefix(value, prefix) || strings.Contains(value, `\`) {
		return fmt.Errorf("%q does not target immutable release %q", value, release)
	}
	relative := strings.TrimPrefix(value, prefix)
	if !safePath(relative) {
		return fmt.Errorf("%q has invalid release path", value)
	}
	if _, found := files["releases/"+release+"/"+relative]; !found {
		return fmt.Errorf("%q is unavailable", value)
	}
	return nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON documents")
		}
		return err
	}
	return nil
}

func encodeCanonical(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func safePath(name string) bool {
	return name != "." && fs.ValidPath(name) && !strings.Contains(name, `\`) && !strings.Contains(strings.Split(name, "/")[0], ":")
}

// Assemble copies a validated bundle into a rooted public assets directory.
// Identical existing immutable files are preserved; different bytes fail.
func Assemble(ctx context.Context, source fs.FS, destination *os.Root) error {
	if destination == nil {
		return errors.New("asset bundle destination is nil")
	}
	if _, err := Validate(source); err != nil {
		return err
	}
	var names []string
	if err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			names = append(names, name)
		}
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		if err := rejectDestinationSymlink(destination, name); err != nil {
			return err
		}
		current, err := destination.ReadFile(name)
		switch {
		case err == nil && bytes.Equal(current, data):
			continue
		case err == nil:
			return fmt.Errorf("asset bundle collision at logical path %q: destination bytes differ from source %q", name, name)
		case !errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf("read destination %q: %w", name, err)
		}
	}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		if err := ensureDestinationDirectories(destination, path.Dir(name)); err != nil {
			return err
		}
		if _, err := destination.Lstat(name); errors.Is(err, fs.ErrNotExist) {
			if err := destination.WriteFile(name, data, 0o644); err != nil {
				return fmt.Errorf("write destination %q: %w", name, err)
			}
		} else if err != nil {
			return fmt.Errorf("inspect destination %q: %w", name, err)
		}
	}
	return nil
}

func ensureDestinationDirectories(root *os.Root, directory string) error {
	if directory == "." {
		return nil
	}
	parts := strings.Split(directory, "/")
	current := ""
	for _, part := range parts {
		current = path.Join(current, part)
		info, err := root.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			if err := root.Mkdir(current, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
				return fmt.Errorf("create destination directory %q: %w", current, err)
			}
			info, err = root.Lstat(current)
		}
		if err != nil {
			return fmt.Errorf("inspect destination directory %q: %w", current, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("destination path %q is not a directory", current)
		}
	}
	return nil
}

func rejectDestinationSymlink(root *os.Root, name string) error {
	parts := strings.Split(name, "/")
	for length := 1; length <= len(parts); length++ {
		current := strings.Join(parts[:length], "/")
		info, err := root.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect destination %q: %w", current, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("destination path %q is a symbolic link", current)
		}
	}
	return nil
}
