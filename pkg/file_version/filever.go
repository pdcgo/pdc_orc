package file_version

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Version struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type FileVersion struct {
	filename       string
	Versions       map[string]*Version `json:"versions"`
	CurrentVersion string              `json:"current_version"`
}

func NewFileVersion(fname string) *FileVersion {
	file := FileVersion{
		filename:       fname,
		Versions:       map[string]*Version{},
		CurrentVersion: "",
	}

	file.Load()

	return &file
}

func (file *FileVersion) Save() error {
	metafname := file.filename + ".metaversion"
	f, err := os.OpenFile(metafname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(file)
	if err != nil {
		return err
	}

	return nil
}

func (file *FileVersion) Load() {
	metafname := file.filename + ".metaversion"
	_, err := os.Stat(metafname)
	if os.IsNotExist(err) {
		err = file.Save()
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	data, err := os.ReadFile(metafname)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, file)
	if err != nil {
		log.Fatal(err)
	}
}

func (file *FileVersion) PathVersion(name string) string {
	return file.filename + "." + GetID(name)
}

func (file *FileVersion) ActivateVersion(name string) error {
	key := GetID(name)
	if file.Versions[key] == nil {
		return fmt.Errorf("version %s not found", name)
	}

	err := CopyFile(file.Versions[key].Path, file.filename)
	if err != nil {
		return err
	}
	file.CurrentVersion = name
	return file.Save()

}

func (file *FileVersion) CopyVersionFrom(src, name string) error {
	var err error
	srckey := GetID(src)

	if file.Versions[srckey] == nil {
		err = file.CreateVersion(src)
		if err != nil {
			return err
		}
	}
	ver := Version{
		ID:   GetID(name),
		Name: name,
		Path: file.PathVersion(name),
	}
	err = CopyFile(file.Versions[srckey].Path, ver.Path)
	if err != nil {
		return err
	}

	file.Versions[ver.ID] = &ver
	err = file.Save()

	if err != nil {
		return err
	}

	return nil

}

func (file *FileVersion) CreateVersion(name string) error {

	origin := file.PathVersion("original")
	_, err := os.Stat(origin)
	if os.IsNotExist(err) {
		err := CopyFile(file.filename, origin)
		if err != nil {
			return err
		}
		ver := Version{
			ID:   GetID("original"),
			Name: "original",
			Path: origin,
		}
		file.CurrentVersion = name
		file.Versions[ver.ID] = &ver

		err = file.Save()
		if err != nil {
			return err
		}
	}

	err = CopyFile(origin, file.PathVersion(name))
	if err != nil {
		return err
	}
	ver := Version{
		ID:   GetID(name),
		Name: name,
		Path: file.PathVersion(name),
	}

	file.Versions[ver.ID] = &ver
	err = file.Save()
	return err
}

func GetID(text string) string {
	hash := md5.Sum([]byte(text))
	hashid := hex.EncodeToString(hash[:])

	return hashid[len(hashid)-6 : len(hashid)-1]
}

func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
