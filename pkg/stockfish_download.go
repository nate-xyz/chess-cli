package pkg

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nate-xyz/chess-cli/api"
)

// Downloading Stockfish for people who don't have it: the official release
// from GitHub, checked against the SHA-256 GitHub publishes, unpacked into
// chess-cli's config folder (no system install, no admin rights). Stockfish
// is GPLv3; its licence is kept next to it.

const stockfishReleases = "https://api.github.com/repos/official-stockfish/Stockfish/releases/latest"

var downloadClient = &http.Client{Timeout: 10 * time.Minute}

// stockfishAssetName is the release file for this OS and CPU, e.g.
// "stockfish-linux-x86-64-universal.tar.gz".
func stockfishAssetName() (string, error) {
	arch := map[string]string{"amd64": "x86-64", "arm64": "arm64", "riscv64": "riscv64"}[runtime.GOARCH]
	switch {
	case runtime.GOOS == "darwin":
		return "stockfish-macos-universal.tar.gz", nil
	case runtime.GOOS == "linux" && arch != "":
		return "stockfish-linux-" + arch + "-universal.tar.gz", nil
	case runtime.GOOS == "windows" && (arch == "x86-64" || arch == "arm64"):
		return "stockfish-windows-" + arch + "-universal.zip", nil
	}
	return "", fmt.Errorf("no Stockfish download for %s/%s; install it yourself", runtime.GOOS, runtime.GOARCH)
}

// downloadedStockfishDir is where a downloaded Stockfish lives.
func downloadedStockfishDir() (string, error) {
	dir, err := api.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stockfish"), nil
}

// downloadedStockfish is the path of a previously downloaded Stockfish, or "".
func downloadedStockfish() string {
	dir, err := downloadedStockfishDir()
	if err != nil {
		return ""
	}
	name := "stockfish"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	p := filepath.Join(dir, name)
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

// DownloadStockfish fetches and unpacks the latest Stockfish, reporting
// progress (0-1), and returns the engine's path.
func DownloadStockfish(progress func(done, total int64)) (string, error) {
	assetName, err := stockfishAssetName()
	if err != nil {
		return "", err
	}
	var release struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Size   int64  `json:"size"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	res, err := downloadClient.Get(stockfishReleases)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub: %s", res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return "", err
	}
	var url, digest string
	var size int64
	for _, a := range release.Assets {
		if a.Name == assetName {
			url, digest, size = a.URL, strings.TrimPrefix(a.Digest, "sha256:"), a.Size
		}
	}
	if url == "" {
		return "", fmt.Errorf("%s isn't in Stockfish's latest release", assetName)
	}
	if digest == "" {
		return "", errors.New("GitHub didn't give a checksum for the download, so it can't be verified")
	}

	dir, err := downloadedStockfishDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	archive := filepath.Join(dir, assetName+".part")
	defer os.Remove(archive)
	if err := downloadFile(url, archive, size, digest, progress); err != nil {
		return "", err
	}
	return unpackStockfish(archive, assetName, dir)
}

// downloadFile saves url to dst, checking its SHA-256.
func downloadFile(url, dst string, size int64, wantSHA string, progress func(done, total int64)) error {
	res, err := downloadClient.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s", res.Status)
	}
	if size <= 0 {
		size = res.ContentLength
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 256<<10)
	var done int64
	for {
		n, err := res.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			h.Write(buf[:n])
			done += int64(n)
			if progress != nil {
				progress(done, size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, wantSHA) {
		return errors.New("the download is corrupt (checksum mismatch); try again")
	}
	return f.Close()
}

// unpackStockfish extracts the engine (and its licence) from the archive:
// the engine is the file named after the archive, e.g.
// stockfish/stockfish-linux-x86-64-universal.
func unpackStockfish(archive, assetName, dir string) (string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(assetName, ".tar.gz"), ".zip")
	engineName := "stockfish"
	if strings.HasSuffix(assetName, ".zip") {
		engineName += ".exe"
	}
	enginePath := filepath.Join(dir, engineName)
	found := false
	extract := func(name string, r io.Reader) error {
		switch b := path.Base(name); {
		case b == base || b == base+".exe":
			found = true
			return writeFile(enginePath, r, 0o755)
		case b == "Copying.txt":
			return writeFile(filepath.Join(dir, "Copying.txt"), r, 0o644)
		}
		return nil
	}
	if strings.HasSuffix(assetName, ".zip") {
		z, err := zip.OpenReader(archive)
		if err != nil {
			return "", err
		}
		defer z.Close()
		for _, f := range z.File {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			err = extract(f.Name, rc)
			rc.Close()
			if err != nil {
				return "", err
			}
		}
	} else {
		f, err := os.Open(archive)
		if err != nil {
			return "", err
		}
		defer f.Close()
		gz, err := gzip.NewReader(f)
		if err != nil {
			return "", err
		}
		tr := tar.NewReader(gz)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}
			if hdr.Typeflag == tar.TypeReg {
				if err := extract(hdr.Name, tr); err != nil {
					return "", err
				}
			}
		}
	}
	if !found {
		return "", errors.New("couldn't find the engine in the download")
	}
	return enginePath, nil
}

func writeFile(p string, r io.Reader, mode os.FileMode) error {
	tmp := p + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, io.LimitReader(r, 512<<20)); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// installHint is how to install Stockfish yourself on this system.
func installHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install stockfish"
	case "windows":
		return "winget install stockfish"
	case "linux":
		for _, pm := range []struct{ bin, cmd string }{
			{"apt", "sudo apt install stockfish"}, {"dnf", "sudo dnf install stockfish"},
			{"pacman", "sudo pacman -S stockfish"}, {"zypper", "sudo zypper install stockfish"},
		} {
			if _, err := os.Stat("/usr/bin/" + pm.bin); err == nil {
				return pm.cmd
			}
		}
	}
	return "see stockfishchess.org/download"
}
