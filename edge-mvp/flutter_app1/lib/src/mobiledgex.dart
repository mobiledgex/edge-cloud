// Copyright (c) 2018, MobiledgeX
import 'dart:async';

class CloudletRequest {
  // TODO: Define
  String id;
  String carrier;
  String tower;
  List<double> gpsInfo = new List(2); // Long: 0, Lat: 1
  String appInfo;

  CloudletRequest() {
    gpsInfo[0] = 0.0; // Long (0, +/-180 degrees)
    gpsInfo[1] = 1.0; // Lat (0, +/-90 degrees)
  }
}

class MatchingEngine {

  final String matchingEngineServer = "service.mobiledgex.com/matchingengine/";
  final String matchingEngine = "matchingengine";

  MatchingEngine(); // Empty.

  // CloudletRequest needs info. Call Platform Plugins:


  /// Returns a Map<String, String> of 2 parts of a request: The server, and the
  /// Service API resource
  Future<Map<String, String>> getCloudletURI(CloudletRequest req) async {
    if (req == null) {
      throw new Exception("CloudletRequest object required."
      );
    }
    // Do post to matching engine
    // TODO: Stub map return
    return {
      "server": "ec2-52-3-246-92.compute-1.amazonaws.com",
      "service": "/api/detect"
    };
  }

}