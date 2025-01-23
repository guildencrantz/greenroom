package mkdocs

import "gopkg.in/yaml.v3"

type MkDocs struct {
	SiteName           string    `yaml:"site_name"`
	SiteUrl            string    `yaml:"site_url,omitempty"`
	RepoUrl            string    `yaml:"repo_url,omitempty"`
	RepoName           string    `yaml:"repo_name,omitempty"`
	EditUri            string    `yaml:"edit_uri,omitempty"`
	EditUriTemplaet    string    `yaml:"edit_uri_template,omitempty"`
	SiteDescription    string    `yaml:"site_description,omitempty"`
	SiteAuthor         string    `yaml:"site_author,omitempty"`
	Copyright          string    `yaml:"copyright,omitempty"`
	RemoteBranch       string    `yaml:"remote_branch,omitempty"`
	RemoteName         string    `yaml:"remote_name,omitempty"`
	Nav                []NavItem `yaml:"nav,omitempty"`
	ExcludeDocs        string    `yaml:"exclude_docs,omitempty"`
	DraftDocs          string    `yaml:"draft_docs,omitempty"`
	NotInNav           string    `yaml:"not_in_nav,omitempty"`
	MarkdownExtensions []any     `yaml:"markdown_extensions,omitempty"`
	Plugins            []any     `yaml:"plugins,omitempty"`
}

func NewMkDocs() *MkDocs {
	return &MkDocs{
		Plugins: []any{"techdocs-core"},
	}
}

type NavItem struct {
	Name string
	Path string
}

func (n *NavItem) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(map[string]string{
		n.Name: n.Path,
	})
}
