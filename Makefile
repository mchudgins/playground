#
# Makefile for 'playground'
#
# generates the html describing the API's
#

pkg/cmd/backend/htmlGen/assets.go: pkg/cmd/backend/htmlGen/test.yaml pkg/cmd/backend/htmlGen/defaultHTML.go
	go run main.go htmlGen pkg/cmd/backend/htmlGen/test.yaml >pkg/cmd/backend/htmlGen/apiList.html
	staticfiles -o pkg/cmd/backend/htmlGen/assets.go -exclude '*.yaml,*.go' pkg/cmd/backend/htmlGen

fmt:
	-goimports -w .

run: pkg/cmd/backend/htmlGen/assets.go
	go run main.go

clean: fmt
	@-rm pkg/cmd/backend/htmlGen/assets.go
