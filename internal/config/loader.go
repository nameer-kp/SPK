package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from multiple sources
type Loader struct {
	homeDir    string
	projectDir string
}

// NewLoader creates a config loader
func NewLoader() (*Loader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Loader{
		homeDir:    home,
		projectDir: cwd,
	}, nil
}

// Load resolves configuration from all sources
func (l *Loader) Load(profileName string) (*ResolvedConfig, error) {
	// Load .env file first (lowest priority)
	envPath := filepath.Join(l.projectDir, ".env")
	_ = godotenv.Load(envPath) // Ignore error if not exists

	// Load global config
	globalConfig := l.loadGlobalConfig()

	// Determine which profile to use
	if profileName == "" {
		profileName = os.Getenv("FLOWCLI_PROFILE")
	}
	if profileName == "" {
		profileName = globalConfig.DefaultProfile
	}
	if profileName == "" {
		profileName = "default"
	}

	// Load the profile
	profile := l.loadProfile(profileName)

	// Load project config
	projectConfig := l.loadProjectConfig()

	// Merge variables (priority: project > profile > env)
	variables := make(map[string]string)

	// Start with profile variables
	for k, v := range profile.Variables {
		variables[k] = v
	}

	// Override with project variables
	for k, v := range projectConfig.Variables {
		variables[k] = v
	}

	// Resolve workflows and plugins directories
	workflowsDir := projectConfig.WorkflowsDir
	if workflowsDir == "" {
		workflowsDir = globalConfig.WorkflowsDir
	}
	if workflowsDir == "" {
		workflowsDir = filepath.Join(l.homeDir, ".atem", "workflows")
	}

	pluginsDir := projectConfig.PluginsDir
	if pluginsDir == "" {
		pluginsDir = globalConfig.PluginsDir
	}
	if pluginsDir == "" {
		pluginsDir = filepath.Join(l.homeDir, ".atem", "plugins")
	}

	return &ResolvedConfig{
		Profile:      profile,
		Variables:    variables,
		WorkflowsDir: workflowsDir,
		PluginsDir:   pluginsDir,
	}, nil
}

func (l *Loader) loadGlobalConfig() GlobalConfig {
	path := filepath.Join(l.homeDir, ".atem", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return GlobalConfig{}
	}

	var config GlobalConfig
	_ = yaml.Unmarshal(data, &config)
	return config
}

func (l *Loader) loadProfile(name string) Profile {
	// Try loading from ~/.atem/profiles/
	path := filepath.Join(l.homeDir, ".atem", "profiles", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{Name: name, Variables: make(map[string]string)}
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return Profile{Name: name, Variables: make(map[string]string)}
	}

	if profile.Variables == nil {
		profile.Variables = make(map[string]string)
	}

	return profile
}

func (l *Loader) loadProjectConfig() ProjectConfig {
	path := filepath.Join(l.projectDir, ".atem.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}
	}

	var config ProjectConfig
	_ = yaml.Unmarshal(data, &config)
	return config
}

// ListProfiles returns available profile names
func (l *Loader) ListProfiles() []string {
	profilesDir := filepath.Join(l.homeDir, ".atem", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return []string{"default"}
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".yaml" {
			name := e.Name()[:len(e.Name())-5] // Remove .yaml
			profiles = append(profiles, name)
		}
	}

	if len(profiles) == 0 {
		return []string{"default"}
	}
	return profiles
}

// DiscoverProjects finds all projects from FLOW_PROJECT_DIRS or current directory
func (l *Loader) DiscoverProjects() (*DiscoveredProjects, error) {
	dirs := l.getProjectDirs()

	var projects []Project
	for _, dir := range dirs {
		project, err := l.loadProject(dir)
		if err != nil {
			continue // Skip invalid projects
		}
		projects = append(projects, *project)
	}

	return &DiscoveredProjects{Projects: projects}, nil
}

// getProjectDirs returns project directories from env var or current dir
func (l *Loader) getProjectDirs() []string {
	envDirs := os.Getenv("FLOW_PROJECT_DIRS")
	if envDirs == "" {
		return []string{l.projectDir}
	}

	var dirs []string
	for _, dir := range strings.Split(envDirs, ",") {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		// Expand ~ to home directory
		if strings.HasPrefix(dir, "~") {
			dir = filepath.Join(l.homeDir, dir[1:])
		}
		dirs = append(dirs, dir)
	}

	if len(dirs) == 0 {
		return []string{l.projectDir}
	}
	return dirs
}

// loadProject loads a single project from a directory
func (l *Loader) loadProject(dir string) (*Project, error) {
	path := filepath.Join(dir, ".atem.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	project.Path = dir

	// Default workflows_dir to ./workflows
	if project.WorkflowsDir == "" {
		project.WorkflowsDir = "./workflows"
	}
	// Make relative paths absolute
	if !filepath.IsAbs(project.WorkflowsDir) {
		project.WorkflowsDir = filepath.Join(dir, project.WorkflowsDir)
	}

	// Default image viewer
	if project.Tools.ImageViewer == "" {
		project.Tools.ImageViewer = "chafa"
	}

	return &project, nil
}
