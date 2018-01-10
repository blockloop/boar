GO_FILES:=$(shell grep -irl --exclude-dir vendor --exclude-dir .git --include \*.go 'type .* interface')
NOT_REAL:=$(GOPATH)

test:
	go test -race -count=3 ./...

covertools: ${GOPATH}/bin/cover ${GOPATH}/bin/goveralls ${GOPATH}/bin/gover
.PHONY: covertools

${GOPATH}/bin/cover:
	go get golang.org/x/tools/cmd/cover

${GOPATH}/bin/goveralls:
	go get github.com/mattn/goveralls

${GOPATH}/bin/gover:
	go get github.com/modocache/gover

${GOPATH}/bin/mockgen:
	go get github.com/golang/mock/mockgen
	
cover: covertools
	@go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | grep -v vendor/ | xargs -L 1 sh -c

coverage.html: cover
	@go tool cover -html=.coverprofile -o $@

coverage.bind.html: cover
	@go tool cover -html=./bind/.coverprofile -o $@

bench:
	@go test -bench=. -benchtime=5s ./

mocks_test.go: ${GO_FILES}
	@mockgen -write_package_comment=false -package=boar -destination=$@ \
		github.com/blockloop/boar \
		HTTPError,Context,ResponseWriter,Handler
	@sed -i 's|github.com/blockloop/boar/vendor/||g' $@
	@sed -i 's|boar "github.com/blockloop/boar"||g' $@
	@sed -i 's|\bboar\b\.||g' $@
