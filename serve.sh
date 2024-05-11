#!/bin/sh
readonly templ="$GOPATH/bin/templ"
"${templ}" generate ./vick-ui/internal/components && go run ./vick-ui
