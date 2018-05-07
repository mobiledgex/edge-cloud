import 'dart:async';
import 'dart:io';
import 'dart:convert'; // For JSON

import 'package:flutter/material.dart';
import 'package:sprintf/sprintf.dart';
import 'package:camera/camera.dart';
import 'package:path_provider/path_provider.dart';
import 'package:http/http.dart' as http;
import 'package:http_parser/http_parser.dart';
import 'package:flutter_app1/src/mobiledgex.dart';

class CameraExampleHome extends StatefulWidget {
  @override
  _CameraExampleHomeState createState() {
    return new _CameraExampleHomeState();
  }
}

IconData cameraLensIcon(CameraLensDirection direction) {
  switch (direction) {
    case CameraLensDirection.back:
      return Icons.camera_rear;
    case CameraLensDirection.front:
      return Icons.camera_front;
    case CameraLensDirection.external:
      return Icons.camera;
  }
  throw new ArgumentError('Unknown lens direction');
}

class _CameraExampleHomeState extends State<CameraExampleHome> {
  bool opening = false;
  CameraController controller;
  String imagePath;
  int pictureCount = 0;
  Timer aTimer;

  MatchingEngine matchingEngine;
  CloudletRequest cloudletRequest;

  @override
  void initState() {
    super.initState();
    matchingEngine = new MatchingEngine();
    cloudletRequest = new CloudletRequest();
  }

  @override
  Widget build(BuildContext context) {
    final List<Widget> headerChildren = <Widget>[];

    final List<Widget> cameraList = <Widget>[];

    if (cameras.isEmpty) {
      cameraList.add(const Text('No cameras found'));
    } else {
      for (CameraDescription cameraDescription in cameras) {
        cameraList.add(
          new SizedBox(
            width: 90.0,
            child: new RadioListTile<CameraDescription>(
              title: new Icon(cameraLensIcon(cameraDescription.lensDirection)),
              groupValue: controller?.description,
              value: cameraDescription,
              onChanged: (CameraDescription newValue) async {
                final CameraController tempController = controller;
                controller = null;
                await tempController?.dispose();
                controller =
                new CameraController(newValue, ResolutionPreset.high);
                await controller.initialize();
                setState(() {});
              },
            ),
          ),
        );
      }
    }

    headerChildren.add(new Column(children: cameraList));
    if (controller != null) {
      headerChildren.add(playPauseButton());
    }
    if (imagePath != null) {
      headerChildren.add(imageWidget());
    }

    final List<Widget> columnChildren = <Widget>[];
    columnChildren.add(new Row(children: headerChildren));
    if (controller == null || !controller.value.initialized) {
      columnChildren.add(const Text('Tap a camera'));
    } else if (controller.value.hasError) {
      columnChildren.add(
        new Text('Camera error ${controller.value.errorDescription}'),
      );
    } else {
      columnChildren.add(
        new Expanded(
          child: new Padding(
            padding: const EdgeInsets.all(5.0),
            child: new Center(
              child: new AspectRatio(
                aspectRatio: controller.value.aspectRatio,
                child: new CameraPreview(controller),
              ),
            ),
          ),
        ),
      );
    }
    return new Scaffold(
      appBar: new AppBar(
        title: const Text('Detect Face'),
      ),
      body: new Column(children: columnChildren),
      floatingActionButton: (controller == null)
          ? null
          : new FloatingActionButton(
        child: const Icon(Icons.camera),
        onPressed: controller.value.isStarted ? capture : null,
      ),
    );
  }

  Widget imageWidget() {
    return new Expanded(
      child: new Align(
        alignment: Alignment.centerRight,
        child: new SizedBox(
          child: new Image.file(new File(imagePath)),
          width: 64.0,
          height: 64.0,
        ),
      ),
    );
  }

  Widget playPauseButton() {
    return new FlatButton(
      onPressed: () {
        setState(
              () {
            if (controller.value.isStarted) {
              controller.stop();
            } else {
              controller.start();
            }
          },
        );
      },
      child:
      new Icon(controller.value.isStarted ? Icons.pause : Icons.play_arrow),
    );
  }

  Future<Null> capture() async {
    if (controller.value.isStarted) {
      final Directory tempDir = await getTemporaryDirectory();
      if (!mounted) {
        return;
      }
      final String tempPath = tempDir.path;
      final String path = '$tempPath/picture${pictureCount++}.jpg';
      await controller.capture(path);


      if (!mounted) {
        return;
      }
      setState(
            () {
          imagePath = path;
        },
      );

      // Here, we say, we don't care when, just upload to the server!
      //var url = "http://ec2-52-3-246-92.compute-1.amazonaws.com/api/detect";
      var imageFile = new File(imagePath);
      var bytes = imageFile.readAsBytesSync();

      // Get the correct cloudlet.
      var uri;
      try {
        Map<String, String> uriResult = await matchingEngine.getCloudletURI(cloudletRequest);
        uri = new Uri.http(uriResult['server'], uriResult['service'], { "q": "dart"});
      } catch (e) {
        return;
      }


      var request = new http.MultipartRequest("POST", uri);
      request.fields['user'] = 'foobarrryTomPortMail@gmail.com';
      request.files.add(new http.MultipartFile.fromBytes(
        'image',
        bytes,
        filename: imagePath,
        contentType: new MediaType("image", "jpeg"))
      );
      request.send().then((response) {
        if (response.statusCode == 200) {
          print("Uploaded!");
          //print("Response length: " + sprintf("%d", response.contentLength));
          response.stream.bytesToString().then((string) {
            print("StreamedResponse: " + string.replaceAll("\n", " "));
            Map<String, dynamic> everaifaces = json.decode(string);

            try {
              var detections = everaifaces['faces'];
              if (detections.length > 0) {
                for (var elm in detections) {
                  var box = elm['bounding_box'];
                  print(sprintf("Acceptable: %s", [elm['acceptable'].toString()]));
                  print(sprintf(
                      "Rectangle box: {height: %d, width: %d, x: %d, y: %d}",
                      [box['height'], box['width'], box['x'], box['y']]));
                }
              }
            } catch (e) {
              print(e.toString());
            }

          });
        }
      });


    }
  }
  Future<Null> semiStreamCapture() async {
    const oneSec = const Duration(seconds:1);
    Timer timer = new Timer.periodic(oneSec, (Timer t) => capture);
    setState(
        () {
          aTimer = timer;
        }
    );
  }
}

List<CameraDescription> cameras;

Future<Null> main() async {
  cameras = await availableCameras();
  runApp(new MaterialApp(home: new CameraExampleHome()));
}