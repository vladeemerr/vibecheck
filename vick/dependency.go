package main

import (
	"encoding/json"
	"go/build"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"unicode"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

type goModInfo struct {
	Version string `json:"Version"`
	Time string `json:"Time"`
	Origin struct {
		VCS string `json:"VCS"`
		URL string `json:"URL"`
		Ref string `json:"Ref"`
		Hash string `json:"Hash"`
	} `json:"Origin"`
}

type Dependency struct {
	Version string
	Path string
	Time time.Time
	Origin Origin
	Licenses []string
}

type Origin struct {
	VCS string
	URL string
	Branch string
	CommitHash string
	CommitTime time.Time
	Committers []string
}

var (
	goPkgMod string
	goPkgModCache string
)

func InitDependencyParserGlobals() {
	goPkgMod = os.Getenv("GOMODCACHE")
	if goPkgMod == "" {
		goPkgMod = os.Getenv("GOPATH")
		if goPkgMod == "" {
			goPkgMod = build.Default.GOPATH
		}

		goPkgMod = filepath.Join(goPkgMod, "pkg", "mod")
	}


	goPkgModCache = filepath.Join(goPkgMod, "cache", "download")
}

func convertToGoModNaming(s string) string {
	var length int
	for _, c := range s {
		if unicode.IsUpper(c) {
			length += 2
		} else {
			length += 1
		}
	}

	var i int
	result := make([]rune, length)

	for _, c := range s {
		if unicode.IsUpper(c) {
			result[i] = rune('!')
			result[i + 1] = unicode.ToLower(c)
			i += 2
		} else {
			result[i] = c
			i++
		}
	}

	return string(result)
}

func ParseDependencies(modPath string) map[string]Dependency {
	result := make(map[string]Dependency)

	queue := []string{modPath}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}

		mod, err := modfile.Parse(path, data, nil)
		if err != nil {
			log.Fatalln(err)
		}

		for _, v := range mod.Require {
			if _, ok := result[v.Mod.Path]; ok {
				log.Printf("Skipping %s, already found\n", v.Mod.Path)
				continue
			}

			log.Println(mod.Module.Mod.Path, "module requires", v.Mod.String())

			dirs := strings.Split(convertToGoModNaming(v.Mod.Path), string(filepath.Separator))
			modPath := filepath.Join(goPkgModCache, filepath.Join(dirs...))

			log.Println("-> trying to locate", v.Mod.Version, "in", modPath)

			bestVer := v.Mod.Version
			nextPath := ""

			err := filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				ver, found := strings.CutSuffix(path, ".mod")
				if found {
					ver = filepath.Base(ver)

					if semver.Compare(ver, bestVer) >= 0 {
						bestVer = ver
						nextPath = path
					}
				}

				return nil
			})

			if err != nil {
				if !os.IsNotExist(err) {
					log.Fatalln(err)
				} else {
					log.Printf("Skipping %s, not found\n", v.Mod.Path)
				}
			}

			if nextPath == "" {
				log.Println("->", v.Mod.String(), " has no versions cached? Skipping...")
				continue
			}

			log.Println("-> found best version available", nextPath)

			baseMod, _ := strings.CutSuffix(nextPath, ".mod")

			infoData, err := os.ReadFile(baseMod + ".info")
			if err != nil {
				log.Println("-> info file was not found, skipping...")
				continue
			}

			var info goModInfo
			err = json.Unmarshal(infoData, &info)
			if err != nil {
				log.Fatalln(err)
			}

			if _, ok := result[v.Mod.Path]; !ok {
				time, err := time.Parse(time.RFC3339, info.Time)
				if err != nil {
					log.Fatalln(err)
				}

				branch := strings.TrimPrefix(info.Origin.Ref, "refs/tags/")

				result[v.Mod.Path] = Dependency{
					Version: info.Version,
					Path: filepath.Join(goPkgMod, v.Mod.String()),
					Time: time,
					Origin: Origin{
						VCS: info.Origin.VCS,
						URL: info.Origin.URL,
						Branch: branch,
						CommitHash: info.Origin.Hash,
					},
				}

				queue = append(queue, nextPath)
			}
		}
	}

	return result
}
