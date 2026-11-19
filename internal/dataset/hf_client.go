package dataset

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laymanmax/steam-localization-atlas/internal/logging"
)

const (
	DefaultRepo = "FronkonGames/steam-games-dataset"
	HFAPIBase   = "https://huggingface.co/api/datasets"
	HFDownload  = "https://huggingface.co/datasets"
)

// HFFile represents a file object returned by the Hugging Face API.
type HFFile struct {
	Type string `json:"type"`
	Oid  string `json:"oid"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}

// DatasetSource handles discovery and downloading of dataset files.
type DatasetSource struct {
	repo       string
	destDir    string
	httpClient *http.Client
}

// NewDatasetSource creates a new dataset source.
func NewDatasetSource(repo, destDir string) *DatasetSource {
	if repo == "" {
		repo = DefaultRepo
	}
	return &DatasetSource{
		repo:    repo,
		destDir: destDir,
		httpClient: &http.Client{
			Timeout: 60 * time.Minute, // Large parquet files may take time on slow links
		},
	}
}

// DiscoverParquetFiles queries the Hugging Face repository API for available Parquet files.
func (ds *DatasetSource) DiscoverParquetFiles() ([]HFFile, error) {
	// First check /tree/main/data
	files, err := ds.fetchTree("data")
	if err == nil && len(files) > 0 {
		var parquets []HFFile
		for _, f := range files {
			if f.Type == "file" && strings.HasSuffix(f.Path, ".parquet") {
				parquets = append(parquets, f)
			}
		}
		if len(parquets) > 0 {
			return parquets, nil
		}
	}

	// Fallback to repository root
	files, err = ds.fetchTree("")
	if err != nil {
		return nil, fmt.Errorf("failed to discover repository files: %w", err)
	}

	var parquets []HFFile
	for _, f := range files {
		if f.Type == "file" && strings.HasSuffix(f.Path, ".parquet") {
			parquets = append(parquets, f)
		}
	}

	if len(parquets) == 0 {
		return nil, fmt.Errorf("no parquet files found in repository %s", ds.repo)
	}

	return parquets, nil
}

func (ds *DatasetSource) fetchTree(subpath string) ([]HFFile, error) {
	url := fmt.Sprintf("%s/%s/tree/main", HFAPIBase, ds.repo)
	if subpath != "" {
		url += "/" + subpath
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Steam-Localization-Atlas/1.0")

	resp, err := ds.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HF API returned HTTP %d for %s", resp.StatusCode, url)
	}

	var files []HFFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode HF API response: %w", err)
	}

	return files, nil
}

// ProgressFunc reports download progress.
type ProgressFunc func(downloadedBytes, totalBytes int64)

// DownloadFile downloads a discovered file to the local destination directory.
func (ds *DatasetSource) DownloadFile(file HFFile, progress ProgressFunc) (string, string, error) {
	if err := os.MkdirAll(ds.destDir, 0755); err != nil {
		return "", "", err
	}

	destPath := filepath.Join(ds.destDir, filepath.Base(file.Path))
	downloadURL := fmt.Sprintf("%s/%s/resolve/main/%s", HFDownload, ds.repo, file.Path)

	logging.Info("Downloading %s (%d bytes) from %s", file.Path, file.Size, downloadURL)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Steam-Localization-Atlas/1.0")

	resp, err := ds.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	tmpFile := destPath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", "", err
	}
	defer func() {
		out.Close()
		_ = os.Remove(tmpFile)
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)

	buf := make([]byte, 1024*1024) // 1MB buffer
	var downloaded int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				return "", "", writeErr
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, file.Size)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", "", readErr
		}
	}

	out.Close()

	// Atomically move tmp file to destPath
	if err := os.Rename(tmpFile, destPath); err != nil {
		return "", "", fmt.Errorf("failed to move downloaded file to %s: %w", destPath, err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	logging.Info("Downloaded %s successfully (SHA256: %s)", destPath, checksum)

	return destPath, checksum, nil
}
