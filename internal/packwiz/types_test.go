package packwiz

import (
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestPackTomlStructure(t *testing.T) {
	// Test that PackToml struct can be instantiated with expected fields
	pack := PackToml{
		Name:       "Test Pack",
		Author:     "Test Author",
		Version:    "1.0.0",
		PackFormat: "packwiz:1.1.0",
		McVersion:  "1.20.1",
	}

	// Set nested struct fields
	pack.Index.File = "index.toml"
	pack.Index.HashFormat = "sha256"
	pack.Versions.Minecraft = "1.20.1"

	if pack.Name != "Test Pack" {
		t.Errorf("Expected Name 'Test Pack', got '%s'", pack.Name)
	}

	if pack.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", pack.Author)
	}

	if pack.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", pack.Version)
	}

	if pack.Index.File != "index.toml" {
		t.Errorf("Expected Index.File 'index.toml', got '%s'", pack.Index.File)
	}

	if pack.Versions.Minecraft != "1.20.1" {
		t.Errorf("Expected Versions.Minecraft '1.20.1', got '%s'", pack.Versions.Minecraft)
	}
}

func TestPackTomlSerialization(t *testing.T) {
	// Test TOML serialization and deserialization
	pack := PackToml{
		Name:        "Serialization Test Pack",
		Author:      "Test Author",
		Version:     "2.0.0",
		PackFormat:  "packwiz:1.1.0",
		McVersion:   "1.19.4",
		Description: "A test pack for serialization",
	}

	pack.Index.File = "index.toml"
	pack.Index.HashFormat = "sha256"
	pack.Index.Hash = "abc123def456"
	pack.Versions.Minecraft = "1.19.4"
	pack.Versions.Fabric = "0.14.21"
	pack.Versions.Forge = "45.2.0"
	pack.Options.AcceptableGameVersions = []string{"1.19", "1.19.4"}

	// Serialize to TOML
	tomlData, err := toml.Marshal(pack)
	if err != nil {
		t.Fatalf("Failed to marshal PackToml to TOML: %v", err)
	}

	// Deserialize from TOML
	var deserializedPack PackToml
	if err := toml.Unmarshal(tomlData, &deserializedPack); err != nil {
		t.Fatalf("Failed to unmarshal TOML to PackToml: %v", err)
	}

	// Verify fields
	if deserializedPack.Name != pack.Name {
		t.Errorf("Name mismatch: expected %s, got %s", pack.Name, deserializedPack.Name)
	}

	if deserializedPack.Version != pack.Version {
		t.Errorf("Version mismatch: expected %s, got %s", pack.Version, deserializedPack.Version)
	}

	if deserializedPack.Versions.Fabric != pack.Versions.Fabric {
		t.Errorf("Fabric version mismatch: expected %s, got %s", pack.Versions.Fabric, deserializedPack.Versions.Fabric)
	}

	if len(deserializedPack.Options.AcceptableGameVersions) != len(pack.Options.AcceptableGameVersions) {
		t.Errorf("AcceptableGameVersions length mismatch: expected %d, got %d",
			len(pack.Options.AcceptableGameVersions), len(deserializedPack.Options.AcceptableGameVersions))
	}
}

func TestModTomlStructure(t *testing.T) {
	// Test ModToml struct
	mod := ModToml{
		Name:     "Test Mod",
		Filename: "testmod.jar",
		Side:     "both",
	}

	// Set nested struct fields
	mod.Download.URL = "https://example.com/mod.jar"
	mod.Download.HashFormat = "sha1"
	mod.Download.Hash = "abc123"
	mod.Update.Modrinth.ModID = "test-mod"
	mod.Update.Curseforge.ProjectID = 123456

	if mod.Name != "Test Mod" {
		t.Errorf("Expected Name 'Test Mod', got '%s'", mod.Name)
	}

	if mod.Side != "both" {
		t.Errorf("Expected Side 'both', got '%s'", mod.Side)
	}

	if mod.Download.URL != "https://example.com/mod.jar" {
		t.Errorf("Expected Download.URL 'https://example.com/mod.jar', got '%s'", mod.Download.URL)
	}

	if mod.Update.Modrinth.ModID != "test-mod" {
		t.Errorf("Expected Update.Modrinth.ModID 'test-mod', got '%s'", mod.Update.Modrinth.ModID)
	}
}

func TestModTomlSides(t *testing.T) {
	// Test different mod sides
	sides := []string{"client", "server", "both"}

	for _, side := range sides {
		mod := ModToml{
			Name:     "Test Mod",
			Filename: "testmod.jar",
			Side:     side,
		}

		if mod.Side != side {
			t.Errorf("Expected Side '%s', got '%s'", side, mod.Side)
		}
	}
}

func TestModTomlSerialization(t *testing.T) {
	// Test TOML serialization for ModToml
	mod := ModToml{
		Name:     "Serialization Test Mod",
		Filename: "test-mod-1.0.0.jar",
		Side:     "client",
	}

	mod.Download.URL = "https://cdn.modrinth.com/data/test/versions/1.0.0/test-mod-1.0.0.jar"
	mod.Download.HashFormat = "sha512"
	mod.Download.Hash = "abcdef123456789"
	mod.Update.Modrinth.ModID = "test-mod-id"
	mod.Update.Modrinth.Version = "1.0.0"
	mod.Update.Curseforge.ProjectID = 123456
	mod.Update.Curseforge.FileID = 789012
	mod.Parse.ModID = "test-mod"
	mod.Parse.Path = "mods/test-mod.pw.toml"

	// Serialize to TOML
	tomlData, err := toml.Marshal(mod)
	if err != nil {
		t.Fatalf("Failed to marshal ModToml to TOML: %v", err)
	}

	// Deserialize from TOML
	var deserializedMod ModToml
	if err := toml.Unmarshal(tomlData, &deserializedMod); err != nil {
		t.Fatalf("Failed to unmarshal TOML to ModToml: %v", err)
	}

	// Verify fields
	if deserializedMod.Name != mod.Name {
		t.Errorf("Name mismatch: expected %s, got %s", mod.Name, deserializedMod.Name)
	}

	if deserializedMod.Download.HashFormat != mod.Download.HashFormat {
		t.Errorf("HashFormat mismatch: expected %s, got %s", mod.Download.HashFormat, deserializedMod.Download.HashFormat)
	}

	if deserializedMod.Update.Curseforge.ProjectID != mod.Update.Curseforge.ProjectID {
		t.Errorf("CurseForge ProjectID mismatch: expected %d, got %d",
			mod.Update.Curseforge.ProjectID, deserializedMod.Update.Curseforge.ProjectID)
	}
}

func TestIndexTomlStructure(t *testing.T) {
	// Test IndexToml struct
	index := IndexToml{
		HashFormat: "sha256",
		Files: []struct {
			File     string `toml:"file"`
			Hash     string `toml:"hash"`
			Metafile bool   `toml:"metafile,omitempty"`
		}{
			{File: "mods/test-mod.pw.toml", Hash: "abc123", Metafile: true},
			{File: "mods/another-mod.pw.toml", Hash: "def456", Metafile: true},
			{File: "config/some-config.toml", Hash: "ghi789", Metafile: false},
		},
	}

	if index.HashFormat != "sha256" {
		t.Errorf("Expected HashFormat 'sha256', got '%s'", index.HashFormat)
	}

	if len(index.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(index.Files))
	}

	if index.Files[0].File != "mods/test-mod.pw.toml" {
		t.Errorf("Expected first file 'mods/test-mod.pw.toml', got '%s'", index.Files[0].File)
	}

	if !index.Files[0].Metafile {
		t.Error("Expected first file to be metafile")
	}

	if index.Files[2].Metafile {
		t.Error("Expected third file to not be metafile")
	}
}

func TestIndexTomlSerialization(t *testing.T) {
	// Test TOML serialization for IndexToml
	index := IndexToml{
		HashFormat: "sha512",
		Files: []struct {
			File     string `toml:"file"`
			Hash     string `toml:"hash"`
			Metafile bool   `toml:"metafile,omitempty"`
		}{
			{File: "mods/fabric-api.pw.toml", Hash: "hash1", Metafile: true},
			{File: "mods/sodium.pw.toml", Hash: "hash2", Metafile: true},
			{File: "resourcepacks/pack.zip", Hash: "hash3", Metafile: false},
		},
	}

	// Serialize to TOML
	tomlData, err := toml.Marshal(index)
	if err != nil {
		t.Fatalf("Failed to marshal IndexToml to TOML: %v", err)
	}

	// Deserialize from TOML
	var deserializedIndex IndexToml
	if err := toml.Unmarshal(tomlData, &deserializedIndex); err != nil {
		t.Fatalf("Failed to unmarshal TOML to IndexToml: %v", err)
	}

	// Verify fields
	if deserializedIndex.HashFormat != index.HashFormat {
		t.Errorf("HashFormat mismatch: expected %s, got %s", index.HashFormat, deserializedIndex.HashFormat)
	}

	if len(deserializedIndex.Files) != len(index.Files) {
		t.Errorf("Files length mismatch: expected %d, got %d", len(index.Files), len(deserializedIndex.Files))
	}

	for i, file := range index.Files {
		if i >= len(deserializedIndex.Files) {
			t.Errorf("Missing file at index %d", i)
			continue
		}

		deserializedFile := deserializedIndex.Files[i]
		if deserializedFile.File != file.File {
			t.Errorf("File[%d] name mismatch: expected %s, got %s", i, file.File, deserializedFile.File)
		}

		if deserializedFile.Hash != file.Hash {
			t.Errorf("File[%d] hash mismatch: expected %s, got %s", i, file.Hash, deserializedFile.Hash)
		}

		if deserializedFile.Metafile != file.Metafile {
			t.Errorf("File[%d] metafile mismatch: expected %v, got %v", i, file.Metafile, deserializedFile.Metafile)
		}
	}
}

func TestJSONCompatibility(t *testing.T) {
	// Test that our structs can be used with JSON as well (for API endpoints)
	pack := PackToml{
		Name:       "JSON Test Pack",
		Author:     "JSON Author",
		Version:    "1.0.0",
		PackFormat: "packwiz:1.1.0",
		McVersion:  "1.20.1",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(pack)
	if err != nil {
		t.Fatalf("Failed to marshal PackToml to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var deserializedPack PackToml
	if err := json.Unmarshal(jsonData, &deserializedPack); err != nil {
		t.Fatalf("Failed to unmarshal JSON to PackToml: %v", err)
	}

	// Verify basic fields
	if deserializedPack.Name != pack.Name {
		t.Errorf("JSON Name mismatch: expected %s, got %s", pack.Name, deserializedPack.Name)
	}
}
