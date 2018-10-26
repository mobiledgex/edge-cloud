# Makefile

GRPCTOOLS_VERSION := Grpc.Tools.1.15.0
TEMP_DIR := packages/$(GRPCTOOLS_VERSION)/tmp
CURL_URL := https://www.nuget.org/api/v2/package/Grpc.Tools
#PROTOC := packages/$(GRPCTOOLS_VERSION)/tools/macosx_x64/protoc
PROTOC := protoc
# Run "brew install coreutils" to get "grealpath" for protoc's requirement of absolute paths
ABSOLUTEPATH := $(shell grealpath ../..)
VSPROJECTROOT := MatchingEngineSDK
APPCLIENTOUT := $(VSPROJECTROOT)/MatchingEngineGrpcLibrary

# Host OS platform tools:
MACOS_PLATFORM := macosx_x64
LINUX_PLATFORM := linux_x64
PLATFORM := $(MACOS_PLATFORM)

GETOUTPUT := $(shell mkdir -p $(TEMP_DIR) && cd $(TEMP_DIR) && curl -sL $(CURL_URL) > tmp.zip; unzip tmp.zip && cd .. && cp -r tmp/tools . && rm -rf tmp && cd ../..)
OUTPUT := $(shell chmod 750 packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/protoc)
OUTPUT := $(shell chmod 750 packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin)

SOLUTIONLOG := /tmp/$(VSPROJECTROOT)_out.log

all: packageRestore build

packageRestore:
	cd $(VSPROJECTROOT) && dotnet restore

build:
	@echo Absolute Path for Proto files are: $(ABSOLUTEPATH)
	mkdir -p $(APPCLIENTOUT)
	@echo Generating GRPC source code...
	$(PROTOC) -I$(ABSOLUTEPATH)/d-match-engine/dme-proto/ -I$(ABSOLUTEPATH)/edge-mvp/thrid_party/googleapis --csharp_out $(APPCLIENTOUT) --grpc_out $(APPCLIENTOUT) $(ABSOLUTEPATH)/d-match-engine/dme-proto/*.proto --plugin=protoc-gen-grpc=packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin
	@echo Making Library...
	$(shell MSBuild $(VSPROJECTROOT)/MatchingEngineSDK.sln > $(SOLUTIONLOG))
	# Print unbounded log to screen:
	cat $(SOLUTIONLOG)
