package content

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
)

type Map struct {
	File string `json:"file"`
	Name string `json:"name"`
}

func getMaps(dir string) (result []*Map, err error) {
	err = walk(dir, func(path string, info os.FileInfo, err error) error {
		mp, err := OpenMapPack(path)
		if err != nil {
			return err
		}
		defer mp.Close()

		maps, err := mp.Maps()
		if err != nil {
			return err
		}
		result = append(result, maps...)
		return err
	}, ".pk3")
	return
}

type MapPack struct {
	*os.File
	*zip.Reader

	path string
}

func OpenMapPack(path string) (*MapPack, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, err
	}
	mp := &MapPack{
		File:   f,
		Reader: r,
		path:   path,
	}
	return mp, nil
}

func (m *MapPack) Maps() ([]*Map, error) {
	maps := make([]*Map, 0)
	for _, f := range m.Reader.File {
		if !hasExts(f.Name, ".bsp") {
			continue
		}
		path := filepath.Join(filepath.Base(filepath.Dir(m.path)), filepath.Base(m.path))
		mapName := strings.TrimSuffix(filepath.Base(f.Name), ".bsp")
		maps = append(maps, &Map{File: path, Name: mapName})
	}
	return maps, nil
}
