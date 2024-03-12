package main

import (
//    "fmt"
    "log"
    "strings"
    "go/build"
    "os"
    "io/fs"
    "path/filepath"
    "flag"
    
    "golang.org/x/mod/semver"
    "golang.org/x/mod/modfile"
)

var (
    goPath string
    mods = map[string]string{}
)

func run(path string) {
    modFileData, err := os.ReadFile(path)
    if err != nil {
        log.Fatalln(err)
    }

    modFile, err := modfile.Parse(path, modFileData, nil)
    if err != nil {
        log.Fatalln(err)
    }

    modPath := filepath.Join(goPath, "pkg", "mod", "cache", "download")

    // TODO: Parse "replaces"
    for _, v := range modFile.Require {
        if _, ok := mods[v.Mod.Path]; ok {
            log.Printf("Skipping %s, already found\n", v.Mod.Path)
            continue
        }
        
        log.Println("Module requires", v.Mod.String())
        
        dirs := strings.Split(v.Mod.Path, string(filepath.Separator))
        // TODO:
        modPath := filepath.Join(modPath, filepath.Join(filepath.Join(dirs...)))

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

            if _, ok := mods[v.Mod.Path]; !ok {
                mods[v.Mod.Path] = v.Mod.String()
                run(nextModPath)
            }
        }
    }
}

func main() {
    goPath = os.Getenv("GOPATH")
    if goPath == "" {
        goPath = build.Default.GOPATH
    }

    flag.Parse()

    modulePath, err := os.Getwd()
    if err != nil {
        // TODO: Make `os.Getwd()` not fatal?
        log.Fatalln(err)
    }

    // TODO: Multiple modules to scan
    if flag.NArg() > 0 {
        modulePath = flag.Arg(0)
    }

    log.Println("Scanning for go.mod file in", modulePath)

    moduleModFilePath := filepath.Join(modulePath, "go.mod")

    if _, err := os.Stat(moduleModFilePath); err != nil {
        if os.IsNotExist(err) {
            return
        }

        log.Fatalln(err)
    }

    log.Println("Found", moduleModFilePath)

    run(moduleModFilePath)
}
