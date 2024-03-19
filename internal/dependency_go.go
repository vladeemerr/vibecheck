package internal

import (
    "time"
    "path/filepath"
    "os"
	"go/build"
    "io/fs"
    "log"
    "strings"
    "encoding/json"

    "golang.org/x/mod/modfile"
    "golang.org/x/mod/semver"
)

var (
    goPkgModPath string
)

func InitGoDependencyParser() {
    goPkgModPath = os.Getenv("GOPATH")
    if goPkgModPath == "" {
        goPkgModPath = build.Default.GOPATH
    }

    goPkgModPath = filepath.Join(goPkgModPath, "pkg", "mod")
}

func ParseGoDependencies(modFilePath string) map[string]Dependency {
    result := map[string]Dependency{}

    modCachePath := filepath.Join(goPkgModPath, "cache", "download")

    // TODO: Use channels
    queue := []string{modFilePath}

    for len(queue) > 0 {
        path := queue[0]
        queue = queue[1:]

        modFileData, err := os.ReadFile(path)
        if err != nil {
            log.Fatalln(err)
        }

        modFile, err := modfile.Parse(path, modFileData, nil)
        if err != nil {
            log.Fatalln(err)
        }

        // TODO: Parse "replaces"
        for _, v := range modFile.Require {
            if _, ok := result[v.Mod.Path]; ok {
                log.Printf("Skipping %s, already found\n", v.Mod.Path)
                continue
            }
            
            log.Println(modFile.Module.Mod.Path, "module requires", v.Mod.String())
            
            dirs := strings.Split(v.Mod.Path, string(filepath.Separator))
            // NOTE: Functional programming go brrr
            modPath := filepath.Join(modCachePath, filepath.Join(filepath.Join(dirs...)))

            log.Println("`-- Trying to locate", v.Mod.Version, "in", modPath)

            bestVersion := v.Mod.Version
            nextModPath := ""

            err := filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
                if err != nil {
                    return err
                }

                version, found := strings.CutSuffix(path, ".mod")
                if found {
                    version = filepath.Base(version) 

                    if semver.Compare(version, bestVersion) >= 0 {
                        bestVersion = version
                        nextModPath = path
                    }
                }

                return nil
            })

            if err != nil {
                log.Fatalln(err)
            }

            if nextModPath != "" {
                log.Println("`-- Found best version available", nextModPath)

                baseModName, _ := strings.CutSuffix(nextModPath, ".mod") 
                infoFileData, err := os.ReadFile(baseModName + ".info")
                if err != nil {
                    continue
                }

                var infoData struct {
                    Version string `json:"Version"`
                    Time string `json:"Time"`
                }

                err = json.Unmarshal(infoFileData, &infoData)
                if err != nil {
                    log.Fatalln(err)
                }

                if _, ok := result[v.Mod.Path]; !ok {
                    time, err := time.Parse(time.RFC3339, infoData.Time)
                    if err != nil {
                        log.Fatalln(err)
                    }
                    
                    result[v.Mod.Path] = Dependency{
                        LangGo,
                        filepath.Join(goPkgModPath, v.Mod.String()),
                        time,
                        Origin{},
                    }
                    queue = append(queue, nextModPath)
                }
            }
        }
    }

    return result
}


