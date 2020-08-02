// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/vfsgen"
	"github.com/spf13/cobra"
)

type modTimeFS struct {
	fs http.FileSystem
}

func (fs modTimeFS) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return modTimeFile{f}, nil
}

type modTimeFile struct {
	http.File
}

func (f modTimeFile) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{fi}, nil
}

type modTimeFileInfo struct {
	os.FileInfo
}

func (modTimeFileInfo) ModTime() time.Time { return time.Time{} }

func StripModTime(fs http.FileSystem) http.FileSystem {
	return modTimeFS{fs}
}

func main() {
	cmd := &cobra.Command{
		Use:  "gen",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateDir := filter.Skip(StripModTime(http.Dir(args[0])), func(path string, fi os.FileInfo) bool {
				return !fi.IsDir() && filepath.Ext(path) == ".go"
			})
			return vfsgen.Generate(templateDir, vfsgen.Options{
				BuildTags:    "!dev",
				VariableName: "Files",
				PackageName:  args[1],
				Filename:     filepath.Join(args[1], "zz_generated.static.go"),
			})
		},
	}
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
