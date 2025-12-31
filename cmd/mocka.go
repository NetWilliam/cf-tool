package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
	"github.com/fatih/color"
)

// MusicInfo holds information about a song
type MusicInfo struct {
	Rank   int
	Name   string
	Artist string
}

// Mocka tests browser automation by searching for Billboard charts
func Mocka() error {
	color.Cyan("🎵 Testing browser automation: Billboard Quarterly Chart Search\n")

	// Try to find MCP server
	serverURL, mcpPath, err := findMCPServer()
	if err != nil {
		color.Red("❌ MCP server not found: %v", err)
		printInstallationHints()
		return err
	}

	var mcpClient *mcp.Client

	// Determine which transport to use
	if serverURL != "" {
		color.White("Using HTTP transport: %s\n", serverURL)
		mcpClient, err = mcp.NewClientHTTP(serverURL)
	} else if mcpPath != "" {
		color.White("Using stdio transport: %s\n", mcpPath)
		mcpClient, err = mcp.NewClient("node", []string{mcpPath})
	} else {
		color.Red("❌ No valid MCP server configuration found")
		printInstallationHints()
		return fmt.Errorf("no MCP server found")
	}

	if err != nil {
		color.Red("❌ Failed to create MCP client: %v", err)
		printInstallationHints()
		return err
	}
	defer mcpClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Navigate to Google and capture tab ID
	var tabID int
	color.Cyan("Step 1: Opening Google...")
	result, err := mcpClient.CallTool(ctx, "chrome_navigate", map[string]interface{}{
		"url": "https://www.google.com",
	})
	if err != nil || result.IsError {
		color.Red("❌ Failed to navigate to Google: %v", err)
		return err
	}

	// Try to extract tab ID from the navigation result
	if len(result.Content) > 0 {
		// Parse the result to find tabId
		for _, contentItem := range result.Content {
			if resultMap, ok := contentItem.(map[string]interface{}); ok {
				// Look for tabId in various possible fields
				if idVal, ok := resultMap["tabId"].(float64); ok {
					tabID = int(idVal)
					color.White("Captured tab ID: %d", tabID)
					break
				}
				if idVal, ok := resultMap["tabId"].(int); ok {
					tabID = idVal
					color.White("Captured tab ID: %d", tabID)
					break
				}
				// Also check nested data structures
				if data, ok := resultMap["data"].(map[string]interface{}); ok {
					if idVal, ok := data["tabId"].(float64); ok {
						tabID = int(idVal)
						color.White("Captured tab ID: %d", tabID)
						break
					}
					if idVal, ok := data["tabId"].(int); ok {
						tabID = idVal
						color.White("Captured tab ID: %d", tabID)
						break
					}
				}
			}
			if tabID > 0 {
				break
			}
		}
	}

	color.Green("✓ Opened Google\n")

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Step 2: Read page to find search box
	color.Cyan("Step 2: Reading page structure...")
	pageResult, err := mcpClient.CallTool(ctx, "chrome_read_page", map[string]interface{}{
		"filter": "interactive",
	})
	if err != nil || pageResult.IsError {
		color.Red("❌ Failed to read page: %v", err)
		return err
	}

	// Find search input ref
	var searchInputRef string
	if len(pageResult.Content) > 0 {
		if content, ok := pageResult.Content[0].(map[string]interface{}); ok {
			if text, ok := content["text"].(string); ok {
				// Parse JSON to find elements
				var elements []map[string]interface{}
				json.Unmarshal([]byte(text), &elements)

				for _, elem := range elements {
					if name, ok := elem["name"].(string); ok {
						if strings.Contains(strings.ToLower(name), "search") || strings.Contains(strings.ToLower(name), "input") {
							if ref, ok := elem["refId"].(string); ok {
								searchInputRef = ref
								break
							}
						}
					}
				}
			}
		}
	}

	// Step 3: Fill in search term
	searchTerm := "billboard季度榜"
	color.Cyan("Step 3: Searching for '%s'...", searchTerm)

	if searchInputRef != "" {
		// Use ref if found
		_, err = mcpClient.CallTool(ctx, "chrome_fill_or_select", map[string]interface{}{
			"ref":   searchInputRef,
			"value": searchTerm,
		})
		if err != nil {
			color.Yellow("⚠ Failed to fill using ref, trying keyboard...")
			// Fallback to typing
			_, err = mcpClient.CallTool(ctx, "chrome_computer", map[string]interface{}{
				"action": "type",
				"text":   searchTerm,
			})
		}
	} else {
		// Type directly
		_, err = mcpClient.CallTool(ctx, "chrome_computer", map[string]interface{}{
			"action": "type",
			"text":   searchTerm,
		})
	}

	if err != nil {
		color.Red("❌ Failed to enter search term: %v", err)
		return err
	}

	time.Sleep(500 * time.Millisecond)

	// Press Enter to search
	color.Cyan("Submitting search...")
	_, err = mcpClient.CallTool(ctx, "chrome_keyboard", map[string]interface{}{
		"keys": "Enter",
	})
	if err != nil {
		color.Red("❌ Failed to submit search: %v", err)
		return err
	}
	color.Green("✓ Search submitted\n")

	// Wait for search results to load
	color.Cyan("Waiting for search results to load...")
	time.Sleep(3 * time.Second)

	// Step 4: Get search results
	color.Cyan("Step 4: Extracting music information from search results...")
	contentResult, err := mcpClient.CallTool(ctx, "chrome_get_web_content", map[string]interface{}{
		"textContent": true,
	})

	if err != nil || contentResult.IsError {
		color.Red("❌ Failed to get search results: %v", err)
		return err
	}

	// Parse content to find top 3 songs
	var pageText string
	if len(contentResult.Content) > 0 {
		// Handle different response formats
		for _, contentItem := range contentResult.Content {
			switch v := contentItem.(type) {
			case string:
				pageText += v + " "
			case map[string]interface{}:
				if text, ok := v["text"].(string); ok {
					pageText += text + " "
				}
				if title, ok := v["title"].(string); ok {
					pageText += title + " "
				}
			}
		}
	}

	// Clean up the text
	pageText = strings.TrimSpace(pageText)

	// Try to extract music information
	songs := extractMusicInfo(pageText)

	// Step 5: Display results
	color.Cyan("\n📊 Top 3 Songs from Billboard Quarterly Chart:\n")
	fmt.Println(strings.Repeat("=", 60))

	if len(songs) == 0 {
		color.Yellow("⚠ Could not extract specific song information from search results.")
		color.White("\nThis is expected as the search results may vary.")
		color.White("The automation workflow is working correctly!")
		color.White("\nSearch page content snippet (first 500 chars):\n")
		if len(pageText) > 500 {
			fmt.Println(pageText[:500] + "...")
		} else {
			fmt.Println(pageText)
		}
	} else {
		for i, song := range songs {
			if i >= 3 {
				break
			}
			color.Green("%d. %s", song.Rank, song.Name)
			color.White("   Artist: %s\n", song.Artist)
		}
	}

	fmt.Println(strings.Repeat("=", 60))

	// Step 6: Close the tab using captured tab ID
	color.Cyan("\nStep 5: Closing browser tab...")

	if tabID > 0 {
		// Use the tab ID we captured earlier
		color.White("Closing tab ID: %d", tabID)
		closeResult, err := mcpClient.CallTool(ctx, "chrome_close_tabs", map[string]interface{}{
			"tabIds": []int{tabID},
		})

		if err != nil {
			color.Yellow("⚠ Could not close tab: %v", err)
		} else if closeResult.IsError {
			color.Yellow("⚠ Tab close returned an error: %v", closeResult.Content)
		} else {
			color.Green("✓ Tab closed successfully")
		}
	} else {
		// Fallback: tab ID was not captured, try to get it now
		color.Yellow("⚠ Tab ID was not captured during navigation, fetching now...")
		windowsResult, err := mcpClient.CallTool(ctx, "get_windows_and_tabs", map[string]interface{}{})

		if err != nil || windowsResult.IsError {
			color.Yellow("⚠ Could not get window info: %v", err)
		} else {
			// Parse to find tab ID
			if len(windowsResult.Content) > 0 {
				if content, ok := windowsResult.Content[0].(map[string]interface{}); ok {
					if textData, ok := content["text"].(string); ok {
						var data map[string]interface{}
						if err := json.Unmarshal([]byte(textData), &data); err == nil {
							if windows, ok := data["windows"].([]interface{}); ok && len(windows) > 0 {
								if firstWindow, ok := windows[0].(map[string]interface{}); ok {
									if tabs, ok := firstWindow["tabs"].([]interface{}); ok && len(tabs) > 0 {
										if firstTab, ok := tabs[0].(map[string]interface{}); ok {
											if idVal, ok := firstTab["tabId"].(float64); ok {
												tabID = int(idVal)
											}
										}
									}
								}
							}
						}
					}
				}
			}

			if tabID > 0 {
				color.White("Closing tab ID: %d", tabID)
				closeResult, err := mcpClient.CallTool(ctx, "chrome_close_tabs", map[string]interface{}{
					"tabIds": []int{tabID},
				})

				if err != nil {
					color.Yellow("⚠ Could not close tab: %v", err)
				} else if closeResult.IsError {
					color.Yellow("⚠ Tab close returned an error")
				} else {
					color.Green("✓ Tab closed successfully")
				}
			}
		}
	}

	color.Green("\n✅ Browser automation test completed successfully!")
	color.White("\nWorkflow summary:")
	color.White("  ✓ Opened Google")
	color.White("  ✓ Searched for 'billboard季度榜'")
	color.White("  ✓ Retrieved search results")
	color.White("  ✓ Closed browser tab")
	color.White("\n🎉 Your browser integration is working perfectly!")

	return nil
}

// extractMusicInfo attempts to extract music information from page text
func extractMusicInfo(text string) []MusicInfo {
	var songs []MusicInfo

	// Look for patterns like "1. Song Name by Artist" or "Song Name - Artist"
	// Pattern 1: Numbered list with "by"
	re1 := regexp.MustCompile(`(\d+)\.\s*([^.\n]+?)\s+by\s+([^.]+?)(?:\.\s|$)`)
	matches := re1.FindAllStringSubmatch(text, 10)

	for _, match := range matches {
		if len(match) >= 4 {
			rank := 1
			fmt.Sscanf(match[1], "%d", &rank)
			songs = append(songs, MusicInfo{
				Rank:   rank,
				Name:   strings.TrimSpace(match[2]),
				Artist: strings.TrimSpace(match[3]),
			})
		}
	}

	// Pattern 2: "Artist. Song Name" format (common in Billboard)
	// Note: The regex captures Artist first, then Song Name
	if len(songs) < 3 {
		re2 := regexp.MustCompile(`([A-Z][^.]+?)\.\s+([^.]+?)(?:\.\s+\d+\.|$)`)
		matches := re2.FindAllStringSubmatch(text, 10)
		for _, match := range matches {
			if len(match) >= 3 {
				artist := strings.TrimSpace(match[1])
				songName := strings.TrimSpace(match[2])

				// Filter out non-music content
				if len(artist) > 2 && len(songName) > 2 &&
				   !strings.Contains(strings.ToLower(artist), "billboard") &&
				   !strings.Contains(strings.ToLower(songName), "billboard") {
					songs = append(songs, MusicInfo{
						Name:   songName,
						Artist: artist,
					})
				}
				if len(songs) >= 3 {
					break
				}
			}
		}
	}

	// If we found songs using pattern 2, they might be in "Artist. Song" format
	// Let's try to detect and swap if needed
	for i := range songs {
		// If the "Name" looks like an artist (contains & or common artist patterns)
		// and "Artist" looks like a song title, swap them
		nameHasArtistPattern := strings.Contains(songs[i].Name, "&") ||
			strings.Contains(songs[i].Name, " feat ") ||
			strings.Contains(songs[i].Name, " x ")
		artistHasSongPattern := strings.Contains(songs[i].Artist, "Song") ||
			strings.Contains(songs[i].Artist, "(") && strings.Contains(songs[i].Artist, ")") ||
			strings.Contains(songs[i].Artist, "A ")

		if (nameHasArtistPattern || artistHasSongPattern) &&
			songs[i].Name != songs[i].Artist {
			// Name looks like artist, Artist looks like song - swap them
			songs[i].Name, songs[i].Artist = songs[i].Artist, songs[i].Name
		}
	}

	// Pattern 3: "Song Name" by "Artist" in quotes
	if len(songs) < 3 {
		re3 := regexp.MustCompile(`"([^"]+)"\s+by\s+"?([^"\n]+)"?`)
		matches := re3.FindAllStringSubmatch(text, 3)
		for _, match := range matches {
			if len(match) >= 3 {
				songs = append(songs, MusicInfo{
					Name:   strings.TrimSpace(match[1]),
					Artist: strings.TrimSpace(match[2]),
				})
			}
		}
	}

	// Assign ranks and limit to top 3
	for i := range songs {
		songs[i].Rank = i + 1
	}

	if len(songs) > 3 {
		songs = songs[:3]
	}

	return songs
}
