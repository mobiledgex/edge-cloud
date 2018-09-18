MobiledgeX GRPC Library for Unity
---------------------------------

Experimental support for experimental GRPC C# Unity library.

Prerequisites:
Visual Studio
DotNet 4.7.x
Protoc install
grealpath (If on MacOS: Install Brew, then run "brew install coreutils")
GRPC for Unity Experimental build: https://packages.grpc.io, BuildID --> Unity build (This is not a NuGet package)

Unpack the GRPC Plugin directory contents and drop it into a unity project. 
  - Copy the entire Plugin folder and drop that folder into a Unity Project Asset Folder Assets/Plugins/GRPC.Core, etc.)
  - Open with Visual Studio, the MatchingEngineSDK/MatchingEngineSDK.sln file.
  - Under MatchingEngineGrpcLibrary dependencies --> Right Click, Edit References.
  - Point each entry to your actual GRPC reference location (where you unpacked the 3 libraries).
  - This GRPC for Unity is not yet a NuGet library, it is experimental.

To finally make the GRPC Library Build:

	make all

This deposits the Debug build here:

	MatchingEngineSDK/MatchingEngineGrpcLibrary/bin/Debug/netstandard2.0/MatchingEngineGrpcLibrary.dll


An Unity application using this library will need to specify DotNet 4.x, along with netstandard2.0.

