package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ModrinthProject represents the API response from Modrinth
type ModrinthProject struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Team        string   `json:"team"`
	Categories  []string `json:"categories"`
}

// ModrinthTeamMember represents a team member from Modrinth
type ModrinthTeamMember struct {
	User struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"user"`
	Role string `json:"role"`
}

// CurseForgeModResponse represents the API response from CurseForge
type CurseForgeModResponse struct {
	Data struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Authors []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"authors"`
		Links struct {
			WebsiteURL string `json:"websiteUrl"`
		} `json:"links"`
	} `json:"data"`
}

// ModInfo represents enriched mod information
type ModInfo struct {
	Name     string
	Author   string
	Platform string
	URL      string
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// GetModrinthInfo fetches mod information from Modrinth API
func GetModrinthInfo(modID string) (*ModInfo, error) {
	url := fmt.Sprintf("https://api.modrinth.com/v2/project/%s", modID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "packwiz-wrapper/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var project ModrinthProject
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, err
	}

	// Fetch team information to get the author
	author := "Unknown"
	if project.Team != "" {
		teamMembers, err := getModrinthTeam(project.Team)
		if err == nil && len(teamMembers) > 0 {
			// Collect all authors
			var authors []string
			for _, member := range teamMembers {
				if member.User.Name != "" {
					authors = append(authors, member.User.Name)
				} else if member.User.Username != "" {
					authors = append(authors, member.User.Username)
				}
			}
			if len(authors) > 0 {
				author = strings.Join(authors, ", ")
			}
		}
	}

	return &ModInfo{
		Name:     project.Title,
		Author:   author,
		Platform: "Modrinth",
		URL:      fmt.Sprintf("https://modrinth.com/mod/%s", project.Slug),
	}, nil
}

func getModrinthTeam(teamID string) ([]ModrinthTeamMember, error) {
	url := fmt.Sprintf("https://api.modrinth.com/v2/team/%s/members", teamID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "packwiz-wrapper/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var members []ModrinthTeamMember
	if err := json.Unmarshal(body, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// GetCurseForgeInfo fetches mod information from CurseForge API
// Note: Requires CURSEFORGE_API_KEY environment variable
func GetCurseForgeInfo(projectID int) (*ModInfo, error) {
	apiKey := os.Getenv("CURSEFORGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CURSEFORGE_API_KEY environment variable not set")
	}

	url := fmt.Sprintf("https://api.curseforge.com/v1/mods/%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var modResp CurseForgeModResponse
	if err := json.Unmarshal(body, &modResp); err != nil {
		return nil, err
	}

	author := "Unknown"
	if len(modResp.Data.Authors) > 0 {
		author = modResp.Data.Authors[0].Name
	}

	return &ModInfo{
		Name:     modResp.Data.Name,
		Author:   author,
		Platform: "CurseForge",
		URL:      modResp.Data.Links.WebsiteURL,
	}, nil
}
