package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func getStreamURLFromPlaylist(playlistURL string) (string, error) {
	resp, err := http.Get(playlistURL)
	if err != nil {
		return "", fmt.Errorf("could not fetch playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("playlist request failed with status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "http") {
			return line, nil
		} else if strings.HasPrefix(strings.ToLower(line), "file1=") {
			return strings.SplitN(line, "=", 2)[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading playlist: %w", err)
	}

	return "", fmt.Errorf("no stream URL found in playlist")
}
