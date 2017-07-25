package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

const (
	defaultSiteStorePath  = "/tmp/sites"
	defaultThemeStorePath = "/tmp/themes"
)

// SiteConfig holds the Hugo configuration.
type SiteConfig struct {
	ID       string `toml:"-"`
	ThemeURL string `toml:"-"`
	SitePath string `toml:"-"`

	ThemesDir  string `toml:"themesDir"`
	ContentDir string `toml:"contentDir"`
	LayoutDir  string `toml:"layoutDir"`
	PublishDir string `toml:"publishDir"`

	BaseURL      string `toml:"baseURL"`
	LanguageCode string `toml:"languageCode"`
	Title        string `toml:"title"`
	Theme        string `toml:"theme"`
}

// Site represents a Hugo site, useful methods will be implemented here.
type Site struct {
	Config *SiteConfig
}

// Build triggers a Hugo build.
func (s *Site) Build() (err error) {
	cmd := exec.Command("hugo")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = s.Config.SitePath
	err = cmd.Run()
	if err != nil {
		log.Error(stderr.String())
		return err
	}
	log.Info(stdout.String())
	return nil
}

// SiteStore wraps useful methods for sites.
type SiteStore struct {
	Sites      map[string]*Site
	SitePath   string
	ThemeStore *ThemeStore
}

// NewSiteStore initializes a new site store.
func NewSiteStore(themeStore *ThemeStore) *SiteStore {
	siteStore := &SiteStore{
		Sites:      make(map[string]*Site),
		SitePath:   defaultSiteStorePath,
		ThemeStore: themeStore,
	}
	_, err := ioutil.ReadDir(siteStore.SitePath)
	if err != nil {
		log.Warning("Site store path doesn't exist, creating.")
		os.Mkdir(siteStore.SitePath, 0700)
	}
	return siteStore
}

// Create creates a new site.
func (s *SiteStore) Create(config *SiteConfig) (site *Site, err error) {
	site = &Site{
		Config: config,
	}

	// First check if the theme exists:
	theme := s.ThemeStore.Find(site.Config.Theme)
	if theme == nil {
		log.Errorf("Theme '%s' doesn't exist!", site.Config.Theme)
		if site.Config.ThemeURL == "" {
			log.Error("No theme URL specified, aborting!")
			return nil, err
		}
		log.Infof("Fetching theme '%s'", site.Config.Theme)
		theme, err = s.ThemeStore.Fetch(site.Config.Theme, site.Config.ThemeURL)
		if err != nil {
			log.Infof("Couldn't fetch theme: %s", err.Error())
			return nil, err
		}
		log.Info("Done fetching theme.")
	}

	sitePath := filepath.Join(s.SitePath, site.Config.ID)

	config.ThemesDir = s.ThemeStore.StorePath
	config.ContentDir = filepath.Join(sitePath, "content")
	config.LayoutDir = filepath.Join(sitePath, "layout")
	config.PublishDir = filepath.Join(sitePath, "public")

	config.SitePath = sitePath

	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err = encoder.Encode(site.Config)
	if err != nil {
		return nil, err
	}

	configFilename := fmt.Sprintf("%s.toml", site.Config.ID)
	configPath := filepath.Join(s.SitePath, configFilename)
	err = ioutil.WriteFile(configPath, buf.Bytes(), 0700)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("hugo", "new", "--config", configPath, site.Config.ID)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Error(stderr.String())
		return nil, err
	}

	log.Info("Output:")
	log.Info(stdout.String())

	newConfigPath := filepath.Join(s.SitePath, site.Config.ID, "config.toml")
	os.Rename(configPath, newConfigPath)

	return site, nil
}

// Find returns the site associated with the given ID.
func (s *SiteStore) Find(id string) (site *Site) {
	sitePath := filepath.Join(s.SitePath, id)
	_, err := ioutil.ReadDir(sitePath)
	if err != nil {
		log.Errorf("Site path '%s' doesn't exist", sitePath)
		return nil
	}
	configPath := filepath.Join(sitePath, "config.toml")
	var config SiteConfig
	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Errorf("Couldn't decode TOML! %s", err.Error())
		return nil
	}
	config.SitePath = sitePath
	site = &Site{
		Config: &config,
	}
	return site
}

// Theme represents a Hugo theme.
type Theme struct{}

// ThemeStore wraps useful methods for looking up, fetching and syncing themes.
type ThemeStore struct {
	Themes    map[string]*Theme
	StorePath string
}

// NewThemeStore initializes a theme store.
func NewThemeStore() *ThemeStore {
	log.Info("Initializing the theme store.")

	themeStore := &ThemeStore{
		StorePath: defaultThemeStorePath,
		Themes:    make(map[string]*Theme),
	}

	dirs, err := ioutil.ReadDir(themeStore.StorePath)
	if err != nil {
		log.Warning("Theme store path doesn't exist, creating.")
		os.Mkdir(themeStore.StorePath, 0700)
	}
	for _, d := range dirs {
		t := &Theme{}
		themeStore.Themes[d.Name()] = t
	}

	if len(dirs) == 0 {
		log.Info("No themes found.")
	}

	return themeStore
}

// Find finds a Hugo theme with the specified name.
func (s *ThemeStore) Find(name string) (theme *Theme) {
	themePath := filepath.Join(s.StorePath, name)
	_, err := ioutil.ReadDir(themePath)
	if err != nil {
		return nil
	}
	t := &Theme{}
	return t
}

// Fetch fetches a theme.
func (s *ThemeStore) Fetch(name string, url string) (theme *Theme, err error) {
	themePath := filepath.Join(s.StorePath, name)
	cmd := exec.Command("git", "clone", url, themePath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Error(stderr.String())
		return nil, err
	}
	t := &Theme{}
	return t, nil
}

// Render generates the pages.
func (s *Site) Render() (err error) {
	return nil
}

// GenerateBundle zips the generated pages and provides a buffer to the compressed content.
func (s *Site) GenerateBundle() ([]byte, error) {
	return []byte(""), nil
}

func main() {
	log.Info("Hugo Site Manager")

	log.Info("Initializing the stores")
	themeStore := NewThemeStore()
	siteStore := NewSiteStore(themeStore)

	log.Info("Creating a sample site")

	siteConfig := &SiteConfig{
		ID:           "myorgid",
		ThemeURL:     "https://github.com/budparr/gohugo-theme-ananke.git",
		BaseURL:      "http://localhost",
		LanguageCode: "en-us",
		Title:        "Test Site",
		Theme:        "ananke",
	}

	// site, _ := siteStore.Create(siteConfig)

	site := siteStore.Find(siteConfig.ID)
	site.Build()
}
