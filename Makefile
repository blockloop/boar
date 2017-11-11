
test:
	go test -coverprofile=.coverprofile .

covertools: ${GOPATH}/bin/cover ${GOPATH}/bin/goveralls ${GOPATH}/bin/gover
.PHONY: covertools

${GOPATH}/bin/cover:
	go get golang.org/x/tools/cmd/cover
${GOPATH}/bin/goveralls :
	go get github.com/mattn/goveralls
${GOPATH}/bin/gover:
	go get github.com/modocache/gover
	
cover: covertools
	@go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | grep -v vendor/ | xargs -L 1 sh -c
