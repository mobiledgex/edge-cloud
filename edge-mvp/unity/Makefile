# Makefile

ABSPATH :=$(shell grealpath ../..)
GRPCTOOLS_VERSION := Grpc.Tools.1.16.0
TEMP_DIR := packages/$(GRPCTOOLS_VERSION)/tmp
CURL_URL := https://www.nuget.org/api/v2/package/Grpc.Tools
PROTOC := protoc
# Run "brew install coreutils" to get "grealpath" for protoc's requirement of absolute paths
ABSOLUTEPATH := $(shell grealpath ../..)
PROTOS_PATH = $(ABSOLUTEPATH)/d-match-engine/dme-proto
THIRD_PARTY_PROTOS_PATH = $(ABSOLUTEPATH)/edge-mvp/third_party

PROTO_INCLUDE_DIRECTORIES := $(PROTOS_PATH) $(THIRD_PARTY_PROTOS_PATH)/googleapis

VSPROJECTROOT := MatchingEngineSDK
APPCLIENTOUT := $(VSPROJECTROOT)/MatchingEngineGrpcLibrary

PROTO_INCLUDE_FLAGS += $(addprefix --proto_path ,$(PROTO_INCLUDE_DIRECTORIES))

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
	$(PROTOC) $(PROTO_INCLUDE_FLAGS) --csharp_out $(APPCLIENTOUT) --grpc_out $(APPCLIENTOUT) $(ABSOLUTEPATH)/d-match-engine/dme-proto/*.proto --plugin=protoc-gen-grpc=packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin
	mkdir -p $(APPCLIENTOUT)/google/api
	$(PROTOC) $(PROTO_INCLUDE_FLAGS) --csharp_out $(APPCLIENTOUT)/google/api --grpc_out $(APPCLIENTOUT) $(ABSOLUTEPATH)/edge-mvp/third_party/googleapis/google/api/*.proto --plugin=protoc-gen-grpc=packages/$(GRPCTOOLS_VERSION)/tools/$(PLATFORM)/grpc_csharp_plugin
	@echo Making Library...
	$(shell MSBuild $(VSPROJECTROOT)/MatchingEngineSDK.sln > $(SOLUTIONLOG))
	# Print unbounded log to screen:
	cat $(SOLUTIONLOG)
