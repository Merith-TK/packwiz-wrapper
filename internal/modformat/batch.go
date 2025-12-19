package modformat

import (
	"sync"
	"time"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

// ModAuthorInfo holds cached author information
type ModAuthorInfo struct {
	Author   string
	Platform string
}

// FetchModAuthorsLight fetches author information with light concurrency control
// Uses a conservative approach: max 3 concurrent requests with 100ms delay between batches
func FetchModAuthors(mods []packwiz.ModToml) map[int]ModAuthorInfo {
	results := make(map[int]ModAuthorInfo)
	resultsMutex := sync.Mutex{}

	const (
		maxConcurrent = 3                      // Conservative: only 3 concurrent requests
		batchDelay    = 100 * time.Millisecond // 100ms between batches
	)

	// Process in small batches to avoid rate limiting
	for i := 0; i < len(mods); i += maxConcurrent {
		end := i + maxConcurrent
		if end > len(mods) {
			end = len(mods)
		}

		batch := mods[i:end]
		var wg sync.WaitGroup

		for idx := range batch {
			globalIdx := i + idx
			mod := batch[idx]

			wg.Add(1)
			go func(index int, m packwiz.ModToml) {
				defer wg.Done()

				platform := GetPlatform(m)
				if platform == "" {
					return
				}

				author := GetModAuthor(m, platform)
				if author != "" {
					resultsMutex.Lock()
					results[index] = ModAuthorInfo{
						Author:   author,
						Platform: platform,
					}
					resultsMutex.Unlock()
				}
			}(globalIdx, mod)
		}

		wg.Wait()

		// Add delay between batches (except for the last one)
		if end < len(mods) {
			time.Sleep(batchDelay)
		}
	}

	return results
}
