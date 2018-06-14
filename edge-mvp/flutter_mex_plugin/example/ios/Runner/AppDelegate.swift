import UIKit
import Flutter
import CoreTelephony

@UIApplicationMain
@objc class AppDelegate: FlutterAppDelegate {
  var count:Double = 0.0;

  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplicationLaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)

    let controller : FlutterViewController = window?.rootViewController as! FlutterViewController;
    let batteryChannel = FlutterMethodChannel.init(name: "samples.flutter.io/sysstate",
                                                   binaryMessenger: controller);

    batteryChannel.setMethodCallHandler({
      (call: FlutterMethodCall, result: FlutterResult) -> Void in
      switch (call.method) {
        case "getBatteryLevel":
          self.receiveBatteryLevel(result: result);
          break;
        case "getApp":
          self.getApp(result: result);
          break;
        case "getLocationInfo":
          self.getLocationInfo(result: result);
          break;
        case "getOperatorInfo":
          self.getOperatorInfo(result: result);
          break;
        default:
          result(FlutterMethodNotImplemented);
      }
    });

    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }

  private func getLocationInfo(result: FlutterResult) {
    // TODO
    count += 1.0;

    var strMap = [String : Double]()
    strMap["lat"] = 0.0;
    strMap["long"] = 0.0;
    strMap["horizontal_accuracy"] = 0.0;
    strMap["vertical_accuracy"] = 0.0;
    strMap["altitude"] = 0.0;
    strMap["course"] = 0.0;
    strMap["speed"] = count;
    strMap["timestamp"] = Double(Date().timeIntervalSince1970 * 1000);
    result(strMap);
  }

  private func receiveBatteryLevel(result: FlutterResult) {
    let device = UIDevice.current;
    device.isBatteryMonitoringEnabled = true;
    if (device.batteryState == UIDeviceBatteryState.unknown) {
        result(FlutterError.init(code: "UNAVAILABLE",
                                 message: "Battery info unavailable",
                                 details: nil));
    } else {
        result(Int(device.batteryLevel * 100));
    }
  }

  private func getOperatorInfo(result: FlutterResult) {
    // TODO
    let mob = CTTelephonyNetworkInfo()
    var strMap = [String : String]()
    if let r = mob.subscriberCellularProvider { //creates CTCarrierObject
        strMap["name"] = r.carrierName
        strMap["id"] = nil // IMEI is no longer allowed in Apps.
        strMap["mnc"] = r.mobileNetworkCode
        strMap["mcc"] = r.mobileCountryCode
        result(strMap)
    } else {
        // Empty Map.
        result(strMap)
    }
  }

  private func getApp(result: FlutterResult) {
        var strMap = [String : String]()
        strMap["dev_name"] = "DEV_NAME_NOT_IMPLEMENTED";
        strMap["app_name"] = Bundle.main.object(forInfoDictionaryKey: "CFBundleName") as? String ?? "";
        strMap["version"] = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "";
        result(strMap);
  }
}




