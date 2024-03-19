package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vladeemerr/vibecheck/internal"
)

func main() {
    internal.InitGoDependencyParser()

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

    deps := internal.ParseGoDependencies(moduleModFilePath)

    for k, v := range deps {
        fmt.Printf("◉ %v:\n  %v\n  %v\n", k, v.Path, v.Time)
    }
}
