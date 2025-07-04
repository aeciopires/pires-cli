#---------------------------
#---------------------------
# VARIABLES
#---------------------------
#---------------------------

# Dependencies
YQ_VERSION=v4.45.1

# Only Ubuntu
#SHELL=/usr/bin/bash
# Only MacOS using brew
SHELL=/opt/homebrew/bin/bash

APP_NAME="pires-cli"
GOLANG_SUPPORTED_PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

#----------------------------------------------------------------------------------------------------------




#---------------------------
#---------------------------
# MAIN
#---------------------------
#---------------------------

# References
# https://ryanstutorials.net/bash-scripting-tutorial/bash-input.php
# https://stackoverflow.com/questions/3743793/makefile-why-is-the-read-command-not-reading-the-user-input
# https://stackoverflow.com/questions/60147129/interactive-input-of-a-makefile-variable
# https://makefiletutorial.com/
# https://stackoverflow.com/questions/589276/how-can-i-use-bash-syntax-in-makefile-targets
# https://til.hashrocket.com/posts/k3kjqxtppx-escape-dollar-sign-on-makefiles
# https://stackoverflow.com/questions/5618615/check-if-a-program-exists-from-a-makefile
# https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/


requirements:
REQUIRED_PACKAGES := docker docker-compose go gcloud wget kubectl trivy psql
$(foreach package,$(REQUIRED_PACKAGES),\
	$(if $(shell command -v $(package) 2> /dev/null),$(info Found `$(package)`),$(error Please install `$(package)`)))

prepare:
	make requirements
	YQ_BINARY="yq_$$(go env GOOS)_$$(go env GOARCH)"
#echo "$$YQ_BINARY"
	wget "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/$${YQ_BINARY}" -O pkg/bilib/fileeditor/internalembeds/yq
	chmod +x pkg/bilib/fileeditor/internalembeds/yq
	pkg/pireslib/fileeditor/internalembeds/yq --version
# Install go packages
	go mod download

.ONESHELL:
build:
	make requirements
	make clean

# Install go packages
	go mod download

	for PLATFORM in $(GOLANG_SUPPORTED_PLATFORMS); do
		GOOS=$$(echo $${PLATFORM} | cut -d'/' -f1)
		GOARCH=$$(echo $${PLATFORM} | cut -d'/' -f2)
		BIN_FILE="bin/${APP_NAME}-$${GOOS}-$${GOARCH}"

		echo "Downloading dependencie: YQ for platform: $${GOOS}/$${GOARCH}"
		YQ_BINARY="yq_$${GOOS}_$${GOARCH}"
		echo "$$YQ_BINARY"
		wget "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/$${YQ_BINARY}" -O pkg/pireslib/fileeditor/internalembeds/yq
		chmod +x pkg/pireslib/fileeditor/internalembeds/yq
		echo "YQ version: $$(pkg/pireslib/fileeditor/internalembeds/yq --version)"

		echo "Building $${BIN_FILE}..."
		export CLI_VERSION="$$(go run . -v)"
		echo "$$CLI_VERSION"
		GOOS=$${GOOS} GOARCH=$${GOARCH} \
		CGO_ENABLED=0 go build -o $${BIN_FILE} .
	done

# Generate SBOM (Software Bill of Materials) using Trivy
# References:
#https://www.bytesizego.com/lessons/sbom-golang
#https://medium.com/@krishnaduttpanchagnula/vulnerability-identification-of-images-and-files-using-sbom-with-trivy-23e1a4a5eea4
	echo "Generating SBOM (Software Bill of Materials)..."
	trivy fs --format cyclonedx --output bin/${APP_NAME}.sbom.json .

	echo "Generate SHA256 checksums for all the built artifacts..."
	cd bin
	if [ "$${GOOS}" == "linux" ]; then
		sha256sum * > checksums.txt
	elif [ "$${GOOS}" == "darwin" ]; then
		shasum -a 256 * > checksums.txt
	else
		echo "Unsupported OS: $${GOOS}. Cannot generate checksums."
		exit 1
	fi
	cd -

	echo "Build completed. Binaries are located in the bin/ directory."
	if [ -d "bin" ]; then
		echo "Binaries:"
		ls -l bin/
	else
		echo "No binaries found in bin/ directory."
	fi

clean:
	rm -rf bin/
	mkdir -p bin
