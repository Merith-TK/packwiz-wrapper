package packwiz

// IndexToml represents the structure of a packwiz index.toml file
type IndexToml struct {
	HashFormat string `toml:"hash-format"`
	Files      []struct {
		File     string `toml:"file"`
		Hash     string `toml:"hash"`
		Metafile bool   `toml:"metafile,omitempty"`
	} `toml:"files"`
}

// ModToml represents the structure of a packwiz mod .pw.toml file
type ModToml struct {
	Name     string `toml:"name"`
	Filename string `toml:"filename"`
	Side     string `toml:"side"`
	Download struct {
		URL        string `toml:"url"`
		HashFormat string `toml:"hash-format"`
		Hash       string `toml:"hash"`
	} `toml:"download"`
	Update struct {
		Modrinth struct {
			ModID   string `toml:"mod-id"`
			Version string `toml:"version"`
		} `toml:"modrinth"`
		Curseforge struct {
			FileID    int `toml:"file-id"`
			ProjectID int `toml:"project-id"`
		} `toml:"curseforge"`
	} `toml:"update,omitempty"`
	// Parse is specific to this program
	Parse struct {
		ModID string `toml:"mod-id"`
		Path  string `toml:"path"`
	} `toml:"parse"`
}

// PackToml represents the structure of a packwiz pack.toml file
type PackToml struct {
	Name        string `toml:"name"`
	Author      string `toml:"author"`
	McVersion   string `toml:"mc-version"`
	PackFormat  string `toml:"pack-format"`
	Description string `toml:"description,omitempty"`
	Index       struct {
		File       string `toml:"file"`
		HashFormat string `toml:"hash-format"`
		Hash       string `toml:"hash"`
	} `toml:"index"`
}
