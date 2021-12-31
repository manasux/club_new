package filesystem

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"net/url"

	mimetype "github.com/gabriel-vasile/mimetype"
	"imuslab.com/arozos/mod/filesystem/shortcut"
)

//Structure definations

type FileData struct {
	Filename    string
	Filepath    string
	Realpath    string
	IsDir       bool
	Filesize    int64
	Displaysize string
	ModTime     int64
	IsShared    bool
	Shortcut    *shortcut.ShortcutData //This will return nil or undefined if it is not a shortcut file
}

type TrashedFile struct {
	Filename         string
	Filepath         string
	FileExt          string
	IsDir            bool
	Filesize         int64
	RemoveTimestamp  int64
	RemoveDate       string
	OriginalPath     string
	OriginalFilename string
}

type FileProperties struct {
	VirtualPath    string
	StoragePath    string
	Basename       string
	VirtualDirname string
	StorageDirname string
	Ext            string
	MimeType       string
	Filesize       int64
	Permission     string
	LastModTime    string
	LastModUnix    int64
	IsDirectory    bool
}

//Check if the two file system are identical.
func MatchingFileSystem(fsa *FileSystemHandler, fsb *FileSystemHandler) bool {
	if fsa.Filesystem == fsb.Filesystem {
		return true
	}
	return false
}

func GetFileDataFromPath(vpath string, realpath string, sizeRounding int) FileData {
	fileSize := GetFileSize(realpath)
	displaySize := GetFileDisplaySize(fileSize, sizeRounding)
	modtime, _ := GetModTime(realpath)

	var shortcutInfo *shortcut.ShortcutData = nil
	if filepath.Ext(realpath) == ".shortcut" {
		scd, err := shortcut.ReadShortcut(realpath)
		if err == nil {
			shortcutInfo = scd
		}
	}

	return FileData{
		Filename:    filepath.Base(realpath),
		Filepath:    vpath,
		Realpath:    filepath.ToSlash(realpath),
		IsDir:       IsDir(realpath),
		Filesize:    fileSize,
		Displaysize: displaySize,
		ModTime:     modtime,
		IsShared:    false,
		Shortcut:    shortcutInfo,
	}

}

func CheckMounted(mountpoint string) bool {
	if runtime.GOOS == "windows" {
		//Windows
		//Check if the given folder exists
		info, err := os.Stat(mountpoint)
		if os.IsNotExist(err) {
			return false
		}
		return info.IsDir()
	} else {
		//Linux
		cmd := exec.Command("mountpoint", mountpoint)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return false
		}
		outstring := strings.TrimSpace(string(out))
		if strings.Contains(outstring, " is a mountpoint") {
			return true
		} else {
			return false
		}
	}
}

func MountDevice(mountpt string, mountdev string, filesystem string) error {
	//Check if running under sudo mode and in linux
	if runtime.GOOS == "linux" {
		//Try to mount the file system
		if mountdev == "" {
			return errors.New("Disk with automount enabled has no mountdev value: " + mountpt)
		}

		if mountpt == "" {
			return errors.New("Invalid storage.json. Mount point not given or not exists for " + mountdev)
		}

		//Check if device exists
		if !fileExists(mountdev) {
			//Device driver not exists.
			return errors.New("Device not exists: " + mountdev)
		}
		//Mount the device
		if CheckMounted(mountpt) {
			log.Println(mountpt + " already mounted.")
		} else {
			log.Println("Mounting " + mountdev + "(" + filesystem + ") to " + filepath.Clean(mountpt))
			cmd := exec.Command("mount", "-t", filesystem, mountdev, filepath.Clean(mountpt))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

		//Check if the path exists
		if !fileExists(mountpt) {
			//Mounted but path still not found. Skip this device
			return errors.New("Unable to find " + mountpt)
		}

	} else {
		return errors.New("Unsupported platform")
	}

	return nil
}

func GetFileSize(filename string) int64 {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	// get the size
	return fi.Size()
}

func IsInsideHiddenFolder(path string) bool {
	thisPathInfo := filepath.ToSlash(filepath.Clean(path))
	pathData := strings.Split(thisPathInfo, "/")
	for _, thispd := range pathData {
		if len(thispd) > 0 && thispd[:1] == "." {
			//This path contain one of the folder is hidden
			return true
		}
	}
	return false
}

/*
	Wildcard Replacement Glob, design to hanle path with [ or ] inside.
	You can also pass in normal path for globing if you are not sure.
*/
func WGlob(path string) ([]string, error) {
	files, err := filepath.Glob(path)
	if err != nil {
		return []string{}, err
	}

	if strings.Contains(path, "[") == true || strings.Contains(path, "]") == true {
		if len(files) == 0 {
			//Handle reverse check. Replace all [ and ] with ?
			newSearchPath := strings.ReplaceAll(path, "[", "?")
			newSearchPath = strings.ReplaceAll(newSearchPath, "]", "?")
			//Scan with all the similar structure except [ and ]
			tmpFilelist, _ := filepath.Glob(newSearchPath)
			for _, file := range tmpFilelist {
				file = filepath.ToSlash(file)
				if strings.Contains(file, filepath.ToSlash(filepath.Dir(path))) {
					files = append(files, file)
				}
			}
		}
	}
	//Convert all filepaths to slash
	for i := 0; i < len(files); i++ {
		files[i] = filepath.ToSlash(files[i])
	}
	return files, nil
}

/*
	Get Directory size, require filepath and include Hidden files option(true / false)
	Return total file size and file count
*/
func GetDirctorySize(filename string, includeHidden bool) (int64, int) {
	var size int64 = 0
	var fileCount int = 0
	err := filepath.Walk(filename, func(thisFilename string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if includeHidden {
				//append all into the file count and size
				size += info.Size()
				fileCount++
			} else {
				//Check if this is hidden
				if !IsInsideHiddenFolder(thisFilename) {
					size += info.Size()
					fileCount++
				}

			}

		}
		return err
	})
	if err != nil {
		return 0, fileCount
	}
	return size, fileCount
}

func GetFileDisplaySize(filesize int64, rounding int) string {
	precisionString := "%." + strconv.Itoa(rounding) + "f"
	var bytes float64
	bytes = float64(filesize)

	var kilobytes float64
	kilobytes = (bytes / 1024)
	if kilobytes < 1 {
		return fmt.Sprintf(precisionString, bytes) + "Bytes"
	}
	var megabytes float64
	megabytes = (float64)(kilobytes / 1024)
	if megabytes < 1 {
		return fmt.Sprintf(precisionString, kilobytes) + "KB"
	}
	var gigabytes float64
	gigabytes = (megabytes / 1024)
	if gigabytes < 1 {
		return fmt.Sprintf(precisionString, megabytes) + "MB"
	}
	var terabytes float64
	terabytes = (gigabytes / 1024)
	if terabytes < 1 {
		return fmt.Sprintf(precisionString, gigabytes) + "GB"
	}
	var petabytes float64
	petabytes = (terabytes / 1024)
	if petabytes < 1 {
		return fmt.Sprintf(precisionString, terabytes) + "TB"
	}
	var exabytes float64
	exabytes = (petabytes / 1024)
	if exabytes < 1 {
		return fmt.Sprintf(precisionString, petabytes) + "PB"
	}
	var zettabytes float64
	zettabytes = (exabytes / 1024)
	if zettabytes < 1 {
		return fmt.Sprintf(precisionString, exabytes) + "EB"
	}

	return fmt.Sprintf(precisionString, zettabytes) + "ZB"
}

func DecodeURI(inputPath string) string {
	inputPath = strings.ReplaceAll(inputPath, "+", "{{plus_sign}}")
	inputPath, _ = url.QueryUnescape(inputPath)
	inputPath = strings.ReplaceAll(inputPath, "{{plus_sign}}", "+")
	return inputPath
}

func GetMime(filepath string) (string, string, error) {
	mime, err := mimetype.DetectFile(filepath)
	return mime.String(), mime.Extension(), err
}

func GetModTime(filepath string) (int64, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return -1, err
	}
	statinfo, err := f.Stat()
	if err != nil {
		return -1, err
	}
	f.Close()
	return statinfo.ModTime().Unix(), nil
}
