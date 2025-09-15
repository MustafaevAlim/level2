package downloader

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Downloader struct {
	BasePath string
	Path     string
}

// / пока не нужна
func (d *Downloader) CreateDirectoryFromPath(p []string) error {
	return os.MkdirAll(strings.Join(p, string(os.PathSeparator)), 0755)
}

func (d *Downloader) DownloadFile(basePath string, p string, in io.Reader) (string, error) {
	defaultName := "index.html"
	path := p
	if !strings.Contains(p, basePath) {
		path = d.BasePath + "/" + path
	}

	if strings.HasSuffix(path, string(os.PathSeparator)) || path == "" {
		path = filepath.Join(path, defaultName)
	}

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		return "", err
	}

	return file.Name(), nil

}

func (d *Downloader) GetPath(p string) string {
	if strings.HasPrefix(p, "http://") {
		return strings.TrimPrefix(p, "http://")

	}
	return strings.TrimPrefix(p, "https://")

}

func (d *Downloader) GetBasePath(u string) string {
	path := d.GetPath(u)
	pathBase := strings.Split(path, "/")
	return pathBase[0]
}

func NewDownloader(u string) Downloader {
	d := Downloader{}
	d.BasePath = d.GetBasePath(u)
	d.Path = u
	return d
}
