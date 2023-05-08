package main

var modlist []PackwizToml

type PackwizToml struct {
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
	} `toml:"update"`
	// Parse is specific to this program
	Parse struct {
		ModID string `toml:"mod-id"`
		Path  string `toml:"path"`
	} `toml:"parse"`
}
