package main

import (
    "fmt"
    "log"
    "flag"
    "path/filepath"
    "io/fs"
    "strings"

    "go/token"
    "go/parser"
)

func main() {
    flag.Parse()
    scanDirs := flag.Args()

    if len(scanDirs) == 0 {
        scanDirs = append(scanDirs, ".")
    }

    var files []string

    err := filepath.WalkDir(scanDirs[0], func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if !d.IsDir() && strings.HasSuffix(path, ".go") {
            files = append(files, path)
        }

        return nil
    })

    if err != nil {
        log.Fatalln(err) 
    }

    for _, f := range files {
        fmt.Printf("%v:\n", f)

        var fset token.FileSet
        ast, err := parser.ParseFile(&fset, f, nil, parser.ImportsOnly)

        if err != nil {
            log.Fatalln(err) 
        }

        for _, v := range ast.Imports {
            fmt.Printf("  %v\n", v.Path.Value)
        }
    }
}
