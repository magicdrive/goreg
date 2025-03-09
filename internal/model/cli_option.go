package model

type Config struct {
	Import ImportConfig `toml:"import"`
	Format FormatConfig `toml:"format"`
}

type ImportConfig struct {
	LocalModule        string `toml:"local_module"`
	OrganizationModule string `toml:"organization_module"`
	Order              string `toml:"order"`
}

type FormatConfig struct {
	MinimizeGroup       bool `toml:"minimize_group"`
	SortIncludeAlias    bool `toml:"sort_include_alias"`
	RemoveImportComment bool `toml:"remove_import_comment"`
}

func (c *Config) SetDefaults() {
	if c.Import.Order == "" {
		c.Import.Order = "std,thirdparty,organization,local"
	}
}
