# Makefile

ABSPATH :=$(shell grealpath ../..)
GRPCTOOLS_VERSION := "Grpc.Tools.1.14.2"
TEMP_DIR := "packages/$(GRPCTOOLS_VERSION)/tmp"
CURL_URL := "https://www.nuget.org/api/v2/package/Grpc.Tools"
PROTOC := protoc
# Run "brew install coreutils" to get "grealpath" for protoc's requirement of absolute paths
ABSOLUTEPATH := $(shell grealpath ../..)
VSPROJECTROOT := "MatchingEngineSDK"
APPCLIENTOUT := "$(VSPROJECTROOT)/MatchingEngineLibrary"

MACOSX_PLATFORM := "macosx_x64"
LINUX_PLATFORM := "linux_x64"
PLATFORM := $(MACOSX_PLATFORM)

GETOUTPUT := $(shell mkdir -p $(TEMP_DIR) && cd $(TEMP_DIR) && curl -sL $(CURL_URL) > tmp.zip; unzip tmp.zip && cd .. && cp -r tmp/tools . && rm -rf tmp && cd ../..)
OUTPUT := $(shell chmod 750 packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/protoc)
OUTPUT := $(shell chmod 750 packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin)

all: packageRestore build

packageRestore:
	cd $(VSPROJECTROOT) && dotnet restore

build:
	@echo "Absolute Path for Proto files are: $(ABSOLUTEPATH)"
	rm -rf $(APPCLIENTOUT)
	mkdir $(APPCLIENTOUT)
	$(PROTOC) -I$(ABSOLUTEPATH)/d-match-engine/dme-proto/ --csharp_out $(APPCLIENTOUT) --grpc_out $(APPCLIENTOUT) $(ABSOLUTEPATH)/d-match-engine/dme-proto/*.proto --plugin=protoc-gen-grpc=packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin

