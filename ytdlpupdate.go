package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ─── yt-dlp Download / Update ───────────────────────────────────────────────

// ytDlpURL is the download URL for yt-dlp Linux binary.
const ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"

// srtFixURL is the download URL for the srt_fix plugin.
const srtFixURL = "https://raw.githubusercontent.com/bindestriche/srt_fix/master/yt_dlp_plugins/postprocessor/srt_fix.py"

// pluginRelPath is where the srt_fix plugin is stored relative to the binary dir.
const pluginRelPath = "yt-dlp-plugins/srt_fix/yt_dlp_plugins/postprocessor/srt_fix.py"

// installYtDlp downloads yt-dlp_linux and the srt_fix plugin,
// then makes the binary executable. Equivalent to install.sh.
func installYtDlp(binaryDir string) error {
	// 1. Create plugin directory
	pluginDir := filepath.Join(binaryDir, filepath.Dir(pluginRelPath))
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("create plugin dir: %w", err)
	}

	// 2. Download srt_fix.py
	pluginPath := filepath.Join(binaryDir, pluginRelPath)
	log.Printf("[yt-dlp] downloading srt_fix.py → %s", pluginPath)
	if err := downloadFile(pluginPath, srtFixURL); err != nil {
		return fmt.Errorf("download srt_fix.py: %w", err)
	}

	// 3. Download yt-dlp_linux
	ytPath := filepath.Join(binaryDir, "yt-dlp_linux")
	log.Printf("[yt-dlp] downloading yt-dlp_linux → %s", ytPath)
	if err := downloadFile(ytPath, ytDlpURL); err != nil {
		return fmt.Errorf("download yt-dlp: %w", err)
	}

	// 4. Make executable
	if err := os.Chmod(ytPath, 0755); err != nil {
		return fmt.Errorf("chmod yt-dlp: %w", err)
	}

	log.Printf("[yt-dlp] installation complete: %s", ytPath)
	return nil
}

// downloadFile downloads a URL to a local file path.
func downloadFile(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d from %s", resp.StatusCode, url)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// needsYtDlpInstall checks whether yt-dlp_linux is present and executable.
func needsYtDlpInstall(binaryDir string) bool {
	ytPath := filepath.Join(binaryDir, "yt-dlp_linux")
	_, err := os.Stat(ytPath)
	return os.IsNotExist(err)
}

// ─── API: Update yt-dlp ─────────────────────────────────────────────────────

// handleUpdateYtDlp triggers a fresh download of yt-dlp and the plugin.
// POST /api/update-ytdlp
func (a *App) handleUpdateYtDlp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	binaryDir := getBinaryDir()

	if err := installYtDlp(binaryDir); err != nil {
		log.Printf("[update-ytdlp] error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"ok":    "false",
			"error": err.Error(),
		})
		return
	}

	log.Printf("[update-ytdlp] yt-dlp updated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ok":      "true",
		"message": "yt-dlp updated successfully",
	})
}
