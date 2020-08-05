package content

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	httputil "github.com/criticalstack/quake-kube/internal/util/net/http"
	"github.com/pkg/errors"
)

func CopyAssets(u *url.URL, dir string) error {
	url := strings.TrimSuffix(u.String(), "/")
	files, err := getManifest(url)
	if err != nil {
		return err
	}

	for _, f := range files {
		path := filepath.Join(dir, f.Name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			continue
		}
		data, err := httputil.GetBody(url + fmt.Sprintf("/assets/%d-%s", f.Checksum, f.Name))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
		if strings.HasPrefix(f.Name, "linuxq3ademo") {
			if err := extractDemoPack(path, dir); err != nil {
				return err
			}
		}
		if strings.HasPrefix(f.Name, "linuxq3apoint") {
			if err := extractPointPacks(path, dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func getManifest(url string) ([]*File, error) {
	data, err := httputil.GetBody(url + "/assets/manifest.json")
	if err != nil {
		return nil, err
	}

	files := make([]*File, 0)
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal %s/assets/manifest.json", url)
	}
	return files, nil
}

var gzipMagicHeader = []byte{'\x1f', '\x8b'}

func extractDemoPack(path, dir string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	idx := bytes.Index(data, gzipMagicHeader)
	data = data[idx:]
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(hdr.Name, ".pk3") {
			fmt.Printf("Downloaded %s\n", hdr.Name)
			data, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			path := filepath.Join(dir, "baseq3", filepath.Base(hdr.Name))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractPointPacks(path, dir string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	idx := bytes.Index(data, gzipMagicHeader)
	data = data[idx:]
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(hdr.Name, ".pk3") {
			fmt.Printf("Downloaded %s\n", hdr.Name)
			data, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			path := filepath.Join(dir, hdr.Name)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
