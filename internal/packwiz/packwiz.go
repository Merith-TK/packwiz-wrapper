package packwiz

import (
	// Modules of packwiz
	"github.com/packwiz/packwiz/cmd"
	_ "github.com/packwiz/packwiz/curseforge"
	_ "github.com/packwiz/packwiz/github"
	_ "github.com/packwiz/packwiz/migrate"
	_ "github.com/packwiz/packwiz/modrinth"
	_ "github.com/packwiz/packwiz/settings"
	_ "github.com/packwiz/packwiz/url"
	_ "github.com/packwiz/packwiz/utils"
)

// Run the Packwiz tool from source with the given arguments
func PackwizExecute() {
	cmd.Execute()
}
