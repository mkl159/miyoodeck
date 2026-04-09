package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = SDCard
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, SDCard) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}
	listFiles(w, path)
}

func listFiles(w http.ResponseWriter, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		jsonError(w, "Cannot read directory: "+err.Error(), http.StatusNotFound)
		return
	}

	var files []FileEntry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	jsonOK(w, map[string]interface{}{
		"path":  path,
		"files": files,
	})
}

// handleDownload streams a single file to the client (fix #10)
func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		jsonError(w, "Missing path", http.StatusBadRequest)
		return
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, SDCard) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		jsonError(w, "File not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		jsonError(w, "Cannot open file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ct := mime.TypeByExtension(filepath.Ext(path))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filepath.Base(path)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	io.Copy(w, f)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		jsonError(w, "Missing path", http.StatusBadRequest)
		return
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, SDCard) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}
	if err := os.RemoveAll(path); err != nil {
		jsonError(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted: " + path})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB in memory; larger files go to temp files on disk (safe on Miyoo Mini)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	destDir := r.FormValue("path")
	if destDir == "" {
		destDir = RomsDir
	}
	destDir = filepath.Clean(destDir)
	if !strings.HasPrefix(destDir, SDCard) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		jsonError(w, "Cannot create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var uploaded []string
	for _, fhs := range r.MultipartForm.File {
		for _, fh := range fhs {
			file, err := fh.Open()
			if err != nil {
				continue
			}
			destPath := filepath.Join(destDir, filepath.Base(fh.Filename))
			dst, err := os.Create(destPath)
			if err != nil {
				file.Close()
				continue
			}
			io.Copy(dst, file)
			dst.Close()
			file.Close()
			uploaded = append(uploaded, fh.Filename)
		}
	}

	jsonOK(w, map[string]interface{}{
		"message":  fmt.Sprintf("Uploaded %d file(s)", len(uploaded)),
		"uploaded": uploaded,
	})
}

func handleUnzip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ZipPath string `json:"zip_path"`
		DestDir string `json:"dest_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.ZipPath = filepath.Clean(req.ZipPath)
	if req.DestDir == "" {
		req.DestDir = filepath.Dir(req.ZipPath)
	}
	req.DestDir = filepath.Clean(req.DestDir)

	if !strings.HasPrefix(req.ZipPath, SDCard) || !strings.HasPrefix(req.DestDir, SDCard) {
		jsonError(w, "Access denied", http.StatusForbidden)
		return
	}

	extracted, err := unzipFile(req.ZipPath, req.DestDir)
	if err != nil {
		jsonError(w, "Unzip failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"message":   fmt.Sprintf("Extracted %d files to %s", extracted, req.DestDir),
		"dest":      req.DestDir,
		"extracted": extracted,
	})
}

func unzipFile(src, dest string) (int, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)) {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), 0755)
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}
		io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		count++
	}
	return count, nil
}

func handleSavesBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	savesPath := SavesDir
	for _, alt := range []string{SavesDir, "/mnt/SDCARD/Saves", "/mnt/SDCARD/Saves/CurrentProfile"} {
		if _, err := os.Stat(alt); err == nil {
			savesPath = alt
			break
		}
	}

	filename := fmt.Sprintf("saves_backup_%s.zip", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	zw := zip.NewWriter(w)
	defer zw.Close()

	filepath.Walk(savesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Dir(savesPath), path)
		f, err := zw.Create(rel)
		if err != nil {
			return nil
		}
		src, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer src.Close()
		io.Copy(f, src)
		return nil
	})
}
