# MEX REST Sample Client

This client shows how to use the C++ REST interface to talk to the MobiledgeX
Distributed Matching Enigne.

It has a few dependencies:
* c++11
* libcurl
* nlohmann_json - This is a header ONLY json library

JSON support install instructions are located here: https://github.com/nlohmann/json

It has 3 APIs in use:

RegisterClient, VerifyLocation, and FindCloudlet REST API calls.
