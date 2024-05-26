package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"flag"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/licensecheck"
)

func main() {
	InitDependencyParserGlobals()
	flag.Parse()

	err := sentry.Init(sentry.ClientOptions{
		Dsn: "",
		Debug: false,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s\n", err)
	}

	defer sentry.Flush(2 * time.Second)

	var path string

	if flag.NArg() > 0 {
		path, _ = filepath.Abs(flag.Arg(0))
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Scanning for go.mod file in", path)

	modPath := filepath.Join(path, "go.mod")

	if _, err := os.Stat(modPath); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		log.Printf("go.mod file was not found in %s, exiting\n", path)
		return
	}

	log.Println("Found", modPath)

	ans := ParseDependencies(modPath)

	vickDir := filepath.Join(path, ".vibecheck")
	err = os.Mkdir(vickDir, 0755)
	if err != nil && os.IsNotExist(err) {
		log.Fatalln(err)
	}

	for k, v := range ans {
		gitDir := filepath.Join(vickDir, k)

		{ // @NOTE: Check if gitDir is valid git repository
			gitCmd := exec.Command("git", "--git-dir", gitDir, "rev-parse")
			err := gitCmd.Run()
			if err != nil {
				switch err.(type) {
				case *exec.ExitError:
					// @NOTE: Clone repository if not exists or invalid
					// @TODO: Push cloning to background
					gitCmd = exec.Command("git", "clone", "--quiet", "--verbose", "--bare",
						"--branch", v.Origin.Branch, v.Origin.URL, gitDir)

					log.Printf("Cloning git repository of %v...", k)
					gitOutput, err := gitCmd.CombinedOutput()
					if err != nil {
						log.Println("MALFORMED, SKIPPING FOR NOW!")
						log.Print(string(gitOutput))
						continue
					}
				}
			} else {
				log.Println(k, "git repository exists, skipping")
			}
		}

		{ // @NOTE: Query release commit hash
			gitCmd := exec.Command("git", "--git-dir", gitDir,
				"rev-parse", v.Origin.Branch)
			gitOutput, err := gitCmd.Output()
			if err != nil {
				log.Fatalln(err)
			}

			v.Origin.CommitHash = string(gitOutput[:len(gitOutput) - 1])
		}

		{ // @NOTE: Query release commit date
			gitCmd := exec.Command("git", "--git-dir", gitDir, "log",
				"-1", "--format=%cd", "--date=iso8601-strict")

			gitOutput, err := gitCmd.Output()
			gitOutput = gitOutput[:len(gitOutput) - 1]
			if err != nil {
				log.Fatalln(err)
			}

			v.Origin.CommitTime, err = time.Parse(time.RFC3339, string(gitOutput))
			if err != nil {
				log.Println(err)
			}

			ans[k] = v
		}

		{ // @NOTE: Query number of commiters
			gitCmd := exec.Command("git", "--git-dir", gitDir, "shortlog",
				"--summary", "--committer")
			gitCmd.Stdin = os.Stdin // @NOTE: git shortlog requires valid stdin
			// @TODO: Pass empty pipe to stdin command

			gitOutput, err := gitCmd.Output()
			if err != nil {
				log.Fatalln(err)
			}

			temp := strings.Split(string(gitOutput), "\n")
			temp = temp[0:len(temp) - 1] // @NOTE: Discard last empty item
			commiters := make([]string, len(temp))
			for i, s := range temp {
				// @TODO: Check for size of split result
				commiters[i] = strings.Split(s, "\t")[1]
			}
			v.Origin.Committers = commiters
		}

		{ // @NOTE: Find license text
			var text []byte

			files := []string{
				"LICENSE",
				"COPYING",
			}

			for _, f := range files {
				text, err = os.ReadFile(filepath.Join(v.Path, f))
				// @NOTE: License found
				if err == nil {
					coverage := licensecheck.Scan(text)

					v.Licenses = make([]string, len(coverage.Match))
					for i, m := range coverage.Match {
						v.Licenses[i] = m.ID
					}

					break
				}
			}
		}

		ans[k] = v
	}

	fmt.Println(len(ans), "dependencies found")
	for k, v := range ans {
		fmt.Printf("* %s:\n", k)
		fmt.Printf("-> Version: %v\n", v.Version)
		fmt.Printf("-> Path: %v\n", v.Path)
		fmt.Printf("-> Time: %v\n", v.Time)
		fmt.Printf("-> VCS: %v\n", v.Origin.VCS)
		fmt.Printf("-> URL: %v\n", v.Origin.URL)
		fmt.Printf("-> Hash: %v\n", v.Origin.CommitHash)
		fmt.Printf("-> Branch: %v\n", v.Origin.Branch)
		if len(v.Licenses) > 0 {
			fmt.Printf("-> License: %v\n", v.Licenses[0])
		}
	}

}
