
test:
	go test -race -count=3 ./...

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

coverage.html: cover
	@go tool cover -html=.coverprofile -o $@

coverage.bind.html: cover
	@go tool cover -html=./bind/.coverprofile -o $@

mocks:
	-@rm mock_*.go 2> /dev/null
	@mockery -case=underscore -all -recursive=false -inpkg
	@rename -f 's|.go|_test.go|g' mock_*
	-@dep prune
