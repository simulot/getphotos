package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/simulot/lib/myflag"
	"github.com/simulot/toolbox/giofs"
)

type App struct {
	DeviceMount      myflag.DirName
	Library          myflag.DirName
	LibrarySubFolder TimeEasyTemplate
	ExcludeList      myflag.Strings
	Move             bool
}

func main() {
	ctx := context.Background()
	PictureDirectory, _ := GetXDGDirectory("XDG_PICTURES_DIR")

	app := App{
		Library:     myflag.DirName(PictureDirectory),
		ExcludeList: []string{".trashed", "screenshoot"},
	}
	app.LibrarySubFolder.Set("Photos/%Y/%Y.%m/%Y.%m.%d")

	flag.Var(&app.DeviceMount, "device", "device mount point. leave empty to search all devices mounted")
	flag.Var(&app.Library, "library", "photo libray path")
	flag.Var(&app.LibrarySubFolder, "subfolder", "Photo destination sub folder based on exposure date, accept %Y,%m,%d for year month day, %H,%M,%S for hours minutes secondes")
	flag.Var(&app.ExcludeList, "exclude", "exclude files containing the string. Can be repeated")
	flag.BoolVar(&app.Move, "move", false, "move files")
	err := app.Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func (app *App) Run(ctx context.Context) error {
	err := myflag.Parse("getphotos.ini")
	if err != nil {
		return err
	}

	if app.DeviceMount != "" {
		err = app.ProcessVolume(ctx,
			giofs.MountInfo{
				DeviceName: app.DeviceMount.String(),
				URI:        "file://" + app.DeviceMount.String(),
				Protocole:  "file",
			})
		if err != nil {
			fmt.Println(err)
		}

	} else {
		mounts, err := giofs.MountList()
		if err != nil {
			return fmt.Errorf("can't get media device list: %w", err)
		}
		if len(mounts) == 0 {
			fmt.Println("No media device found")
			return nil
		}
		for _, m := range mounts {
			switch m.Protocole {
			case "mtp", "file", "gphoto2":
				err = app.ProcessVolume(ctx, m)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}

func (app *App) ProcessVolume(ctx context.Context, m giofs.MountInfo) error {
	done := PleaseWait("Exploring " + m.DeviceName)
	defer done()

	fsys := giofs.GIOFS(m.URI)

	p := "."
	switch m.Protocole {
	case "gphoto2", "mtp":
		pp, err := fs.Glob(fsys, "*/DCIM")
		if err != nil {
			return err
		}
		if len(pp) == 0 {
			return nil
		}
		p = pp[0]
	}

	return fs.WalkDir(fsys, p, func(path string, info fs.DirEntry, err error) error {
		done()
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".jpg", ".jpeg", ".heic", ".mp4", ".png":
			err = app.ProcessFile(ctx, fsys, path, info)
		}
		if err != nil {
			fmt.Printf("Can't process '%s': %s\n", path, err)
			if errors.Is(err, fs.ErrNotExist) {
				return err
			}
			err = nil
		}
		return nil
	})
}

var CopyBuffer = make([]byte, 1024*1024)

func (app *App) ProcessFile(ctx context.Context, fsys fs.FS, path string, info fs.DirEntry) error {
	var d time.Time

	lowerPath := strings.ToLower(path)
	for _, b := range app.ExcludeList {
		b = strings.ToLower(b)
		if strings.Contains(lowerPath, b) {
			fmt.Println("Skiping " + path)
			return nil
		}
	}

	progFn, doneFn := Progression("Processing " + path)
	defer doneFn()

	source, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()
	s, err := source.Stat()
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}

	r := io.TeeReader(source, &buffer)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png":
		d, err = GetJPGDateFromExif(r)
	}

	if err != nil || d.IsZero() {
		m := reDate.FindStringSubmatch(filepath.Base(path))
		if len(m) == 3 {
			if len(m[2]) < 9 {
				m[2] = m[2] + strings.Repeat("0", 9-len(m[2]))
			}
			d, err = time.ParseInLocation("20060102T150405999", m[1]+"T"+m[2], time.Local)
		}
	}

	if err != nil || d.IsZero() {
		var i fs.FileInfo
		i, err = info.Info()
		if err == nil {
			d = i.ModTime()
		}
	}
	if err != nil || d.IsZero() {
		d = time.Now()
	}

	sub := app.LibrarySubFolder.Format(d)
	destName := filepath.Join(app.Library.String(), sub, filepath.Base(path))

	err = os.MkdirAll(filepath.Dir(destName), 0750)
	if err != nil {
		return err
	}

	w, err := os.Create(destName)
	if err != nil {
		return err
	}

	pw := NewProgressionWriter(w, int(s.Size()), progFn)
	_, err = io.CopyBuffer(pw, io.MultiReader(&buffer, source), CopyBuffer)
	if err != nil {
		w.Close()
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	if app.Move {
		return Remove(fsys, path)
	}

	return nil
}

var reDate = regexp.MustCompile(`(20\d{6})(?:\D)(\d{6,9})`)

func GetJPGDateFromExif(r io.Reader) (time.Time, error) {
	x, err := exif.Decode(r)
	if err != nil {
		if exif.IsCriticalError(err) {
			return time.Time{}, err
		}
	}
	return x.DateTime()
}
