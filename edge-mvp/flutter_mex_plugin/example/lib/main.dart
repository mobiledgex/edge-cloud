import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mex_plugin/flutter_mex_plugin.dart';

import 'package:sprintf/sprintf.dart';

void main() => runApp(new MyApp());

class MyApp extends StatefulWidget {
  @override
  _MyAppState createState() => new _MyAppState();
}

class _MyAppState extends State<MyApp> {
  static const platform = const MethodChannel("samples.flutter.io/sysstate");

  String _platformVersion = 'Unknown';
  String _batteryLevel = "Unknown battery level.";
  var _app;
  var _operatorInfo;
  var _locationInfo;

  // For visible text updates.
  var _textStateWidget = new Text(
    'Uninitialized',
  );

  @override
  initState() {
    super.initState();
    initPlatformState();
  }

  // Platform messages are asynchronous, so we initialize in an async method.
  initPlatformState() async {
    String platformVersion;

    // Platform messages may fail, so we use a try/catch PlatformException.
    try {
      platformVersion = await FlutterMexPlugin.platformVersion;
    } on PlatformException {
      platformVersion = 'Failed to get platform version.';
    }

    // If the widget was removed from the tree while the asynchronous platform
    // message was in flight, we want to discard the reply rather than calling
    // setState to update our non-existent appearance.
    if (!mounted)
      return;
    await updateInfo();

    setState(() {
      _platformVersion = platformVersion;
    });

  }

  @override
  Widget build(BuildContext context) {
    return new MaterialApp(
      home: new Scaffold(
        appBar: new AppBar(
          title: new Text('MEX Plugin Example app'),
        ),

        body: new Center(
          child: _textStateWidget,
        ),
        floatingActionButton: new FloatingActionButton(
          child: const Icon(Icons.camera),
          onPressed: updateInfo,
        ),
      ),
    );
  }

  Future<Null> updateInfo() async {
    String batteryLevel;
    try {
      final int result = await platform.invokeMethod("getBatteryLevel");
      batteryLevel = 'Battery level at $result % .';
    } on PlatformException catch (e) {
      batteryLevel = "Failed to get battery level: '$e.message}'.";
    }

    var app;
    try {
      final result = await platform.invokeMethod("getApp");
      app = result;
      print(sprintf("App: %s", [result.toString()]));
    } on PlatformException catch (e) {
      app = {};
      print("Failed to get app description: '$e.message}'.");
    }

    var operatorInfo;
    try {
      final result = await platform.invokeMethod("getOperatorInfo");
      operatorInfo = result;
      print(sprintf("operatorInfo: %s", [result.toString()]));
    } on PlatformException catch (e) {
      operatorInfo = {};
      print("Failed to get operatorInfo : '$e.message}'.");
    }

    // This one is a long running async call. Location is not synchronous.
    var locationInfo;
    try {
      final result = await platform.invokeMethod("getLocationInfo");
      locationInfo = result;
      print(sprintf("locationInfo: %s", [result.toString()]));
    } on PlatformException catch (e) {
      locationInfo = {};
      print("Failed to get locationInfo: '$e.message}'.");
    }

    setState(() {
      _batteryLevel = batteryLevel;
      _app = app;
      _operatorInfo = operatorInfo;
      _locationInfo = locationInfo;

      // Update Text Widget.
      _textStateWidget = new Text('Running on: $_batteryLevel\n, $_platformVersion\n, app: $_app\n,'
          'location: $_locationInfo\n, operatorInfo: $_operatorInfo\n');
    });
  }


}
