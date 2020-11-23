package content

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	AssetsDir string
}

func NewRouter(cfg *Config) (*echo.Echo, error) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.BodyLimit("100M"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Map pack upload</title>
</head>
<body>

<a href="/maps">Show maps</a>

<h1>Upload map pack file</h1>

<form action="/maps" method="post" enctype="multipart/form-data">
	GameName: <input type="text" name="name" value="baseq3" /><br>
    Files: <input type="file" name="file"><br><br>
    <input type="submit" value="Submit">
</form>
</body>
</html>`)
	})
	e.GET("/assets/manifest.json", func(c echo.Context) error {
		files, err := getAssets(cfg.AssetsDir)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, files, "   ")
	})
	e.GET("/assets/*", func(c echo.Context) error {
		path := filepath.Join(cfg.AssetsDir, trimAssetName(c.Param("*")))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "file not found")
		}
		return c.File(path)
	})
	e.GET("/maps", func(c echo.Context) error {
		maps, err := getMaps(cfg.AssetsDir)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, maps, "    ")
	})
	e.POST("/maps", func(c echo.Context) error {
		name := c.FormValue("name")
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		if hasExts(file.Filename, ".zip") {
			r, err := zip.NewReader(src, file.Size)
			if err != nil {
				return err
			}
			files := make([]string, 0)
			for _, f := range r.File {
				if !hasExts(f.Name, ".pk3") {
					continue
				}
				pak, err := f.Open()
				if err != nil {
					return err
				}
				defer pak.Close()

				dst, err := os.Create(filepath.Join(cfg.AssetsDir, name, filepath.Base(f.Name)))
				if err != nil {
					return err
				}
				defer dst.Close()

				if _, err = io.Copy(dst, pak); err != nil {
					return err
				}
				files = append(files, filepath.Base(f.Name))
			}
			if len(files) == 0 {
				return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s did not contain any map pack files.</p>", file.Filename))
			}
			for i, _ := range files {
				files[i] = "<li>" + files[i] + "</li>"
			}
			return c.HTML(http.StatusOK, fmt.Sprintf("<p>Loaded the following map packs from file %s:</p><ul>%s</ul>", file.Filename, strings.Join(files, "")))
		}
		dst, err := os.Create(filepath.Join(cfg.AssetsDir, name, file.Filename))
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return err
		}
		return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully.</p>", filepath.Join(name, file.Filename)))
	})
	return e, nil
}

// trimAssetName returns a path string that has been prefixed with a crc32
// checksum.
func trimAssetName(s string) string {
	d, f := filepath.Split(s)
	f = f[strings.Index(f, "-")+1:]
	return filepath.Join(d, f)
}
