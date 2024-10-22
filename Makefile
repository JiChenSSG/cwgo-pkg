TOOLS_SHELL="./hack/tools.sh"

.PHONY: test
test:
	go test -race -covermode=atomic -coverprofile=coverage.txt ./registry/nacos/nacoskitex/v2/...



.PHONY: vet
vet:
	chmod +x ${TOOLS_SHELL}
	@${TOOLS_SHELL} vet
	@echo "vet check finished"