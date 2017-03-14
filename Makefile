#
# Makefile for 'playground'
#
# generates the html describing the API's
#

pkg/cmd/backend/htmlGen/assets.go: pkg/cmd/backend/htmlGen/test.yaml pkg/cmd/backend/htmlGen/defaultHTML.go
	go run main.go htmlGen pkg/cmd/backend/htmlGen/test.yaml >pkg/cmd/backend/htmlGen/apiList.html
	staticfiles -o pkg/cmd/backend/htmlGen/assets.go -exclude '*.yaml,*.go' pkg/cmd/backend/htmlGen

pkg/cmd/backend/assets.go: pkg/cmd/backend/assets/service.swagger.json
	staticfiles -o pkg/cmd/backend/assets.go pkg/cmd/backend/assets

fmt:
	-goimports -w .

run: pkg/cmd/backend/htmlGen/assets.go pkg/cmd/backend/assets.go
	go run main.go

clean: fmt
	@-rm pkg/cmd/backend/htmlGen/assets.go
