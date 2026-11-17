package pkg

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUnpackStockfishTar(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	add := func(name, body string) {
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg})
		tw.Write([]byte(body))
	}
	add("stockfish/README.md", "readme")
	add("stockfish/Copying.txt", "GPL")
	add("stockfish/stockfish-linux-x86-64-universal", "ENGINE")
	add("../../escape", "nope") // must not be written anywhere
	tw.Close()
	gz.Close()
	archive := filepath.Join(dir, "sf.tar.gz")
	os.WriteFile(archive, buf.Bytes(), 0o644)

	path, err := unpackStockfish(archive, "stockfish-linux-x86-64-universal.tar.gz", dir)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "ENGINE" {
		t.Errorf("engine = %q", b)
	}
	if st, _ := os.Stat(path); runtime.GOOS != "windows" && st.Mode().Perm()&0o100 == 0 { // no executable bit on Windows
		t.Error("engine isn't executable")
	}
	if _, err := os.Stat(filepath.Join(dir, "Copying.txt")); err != nil {
		t.Error("licence not kept")
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "..", "escape")); err == nil {
		t.Error("archive wrote outside its folder")
	}
}

func TestUnpackStockfishZip(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("stockfish/stockfish-windows-x86-64-universal.exe")
	w.Write([]byte("EXE"))
	zw.Close()
	archive := filepath.Join(dir, "sf.zip")
	os.WriteFile(archive, buf.Bytes(), 0o644)
	path, err := unpackStockfish(archive, "stockfish-windows-x86-64-universal.zip", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "stockfish.exe") {
		t.Errorf("path %s", path)
	}
}

func TestDownloadChecksum(t *testing.T) {
	body := []byte("the real engine")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(body) }))
	defer srv.Close()
	sum := sha256.Sum256(body)
	dst := filepath.Join(t.TempDir(), "f")
	if err := downloadFile(srv.URL, dst, int64(len(body)), hex.EncodeToString(sum[:]), nil); err != nil {
		t.Fatalf("good download rejected: %v", err)
	}
	if err := downloadFile(srv.URL, dst, int64(len(body)), strings.Repeat("0", 64), nil); err == nil {
		t.Error("corrupt download accepted")
	}
}

func TestStockfishAssetName(t *testing.T) {
	name, err := stockfishAssetName()
	if err != nil {
		t.Skip(err)
	}
	if !strings.HasPrefix(name, "stockfish-") || !(strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")) {
		t.Errorf("asset %q", name)
	}
}
