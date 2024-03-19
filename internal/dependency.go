package internal

import (
	"time"
)

type Lang int

const (
    LangUndefined Lang = iota
    LangGo
)

type Vcs int

const (
    VcsUndefined Vcs = iota
    VcsGit
)

type Origin struct {
    Vcs Vcs
    Url string
    Ref string
}

type Dependency struct {
    Lang Lang
    Path string
    Time time.Time
    Origin Origin
}

func ParseVcs(name string) Vcs {
    switch name {
    case "git":
        return VcsGit
    }

    return VcsUndefined
}
