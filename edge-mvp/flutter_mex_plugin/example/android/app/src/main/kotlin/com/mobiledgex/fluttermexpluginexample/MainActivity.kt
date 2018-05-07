package com.mobiledgex.fluttermexpluginexample

import android.os.Bundle

import io.flutter.app.FlutterActivity
import io.flutter.plugins.GeneratedPluginRegistrant

import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import io.flutter.plugin.common.MethodChannel.MethodCallHandler
import io.flutter.plugin.common.MethodChannel.Result

import android.content.Context
import android.content.ContextWrapper
import android.Manifest
import android.content.Intent
import android.content.IntentFilter
import android.os.BatteryManager
import android.os.Build.VERSION
import android.os.Build.VERSION_CODES

import android.content.pm.PackageManager

class MainActivity(): FlutterActivity() {
  private val CHANNEL = "samples.flutter.io/sysstate"

  val APP_ON_PERMISSION_RESULT_READ_PHONE_STATE = 1
  var count: Double = 0.0

  override fun onCreate(savedInstanceState: Bundle?) {
    super.onCreate(savedInstanceState)
    GeneratedPluginRegistrant.registerWith(this)

    doCheckPermissions()

    // Flutter's StandardMessageCodec defines allowable types through this interface:
    // https://docs.flutter.io/flutter/services/StandardMessageCodec-class.html
    MethodChannel(flutterView, CHANNEL).setMethodCallHandler { call, result ->
      when (call.method) {
        "getBatteryLevel" -> {
          val batteryLevel = getBatteryLevel()
          if (batteryLevel != -1) {
            result.success(batteryLevel)
          } else {
            result.error("UNAVAILABLE", "Battery level not available.", null)
          }
        }
        "getApp" -> {
          var info = getApp()
          if (info != null) {
            result.success(info)
          } else {
            result.error("UNAVAILABLE", "Cannot obtain application info.", null)
          }
        }
        "getLocationInfo" -> {
          var info = getLocationInfo()
          if (info != null) {
            result.success(info)
          } else {
            result.error("UNAVAILABLE", "Cannot obtain device GPS location info.", null)
          }
        }
        "getOperatorInfo" -> {
          var info = getOperatorInfo()
          if (info != null) {
            result.success(info)
          } else {
            result.error("UNAVAILABLE", "Cannot obtain operator info.", null)
          }
        }
        else -> {
          result.notImplemented()
        }
      }
    }
  }

  // Placement: "getBatteryLevel" is in "example", and not part of the plugin folder structure?
  private fun getBatteryLevel(): Int {
    val batteryLevel: Int
    if (VERSION.SDK_INT >= VERSION_CODES.LOLLIPOP) {
      val batteryManager = getSystemService(Context.BATTERY_SERVICE) as BatteryManager
      batteryLevel = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY)
    } else {
      val intent = ContextWrapper(applicationContext).registerReceiver(null, IntentFilter(Intent.ACTION_BATTERY_CHANGED))
      batteryLevel = intent!!.getIntExtra(BatteryManager.EXTRA_LEVEL, -1) * 100 / intent.getIntExtra(BatteryManager.EXTRA_SCALE, -1)
    }

    return batteryLevel
  }

  private fun getOperatorInfo(): Map<String, String> {
    // TODO:
    // ID is of type string, and redundant to ID?
    val telephonyManager = getSystemService(android.content.Context.TELEPHONY_SERVICE)
            as android.telephony.TelephonyManager

    // https://developer.android.com/reference/android/telephony/TelephonyManager.html
    val ops = mapOf(
      "name" to telephonyManager.getNetworkOperatorName() as String,
      "id" to telephonyManager.getDeviceId() as String, // IMEI
      "mnc" to telephonyManager.getNetworkOperator() as String, // Mnc + mcc. Hm...
      "mcc" to telephonyManager.getNetworkCountryIso() as String
    )
    return ops
  }

  private fun getApp(): Map<String, String> {
    val appInfo: android.content.pm.ApplicationInfo = applicationContext.getApplicationInfo()
    val stringId: Int = applicationInfo.labelRes
    val app_name: String = if (stringId == 0) applicationInfo.nonLocalizedLabel.toString() else applicationContext.getString(stringId)

    val pInfo: android.content.pm.PackageInfo = this.getPackageManager().getPackageInfo(getPackageName(), 0)
    val version: Long = pInfo.versionCode.toLong(); // Android P: Long value.

    val app = mapOf(
      "dev_name" to "DEV_NAME_NOT_IMPLEMENTED",
      "app_name" to app_name,
      "version" to version.toString()
    )
    return app // App identifier
  }

  // Should call from an async context.
  private fun getLocationInfo(): Map<String, Double> {
    count += 1.0
    val gps = mapOf(
      "lat" to 0.0,
      "long" to 0.0,
      "horizontal_accuracy" to 0.0,
      "vertical_accuracy" to 0.0,
      "altitude" to 0.0,
      "course" to 0.0,
      "speed" to count,
      "timestamp" to java.lang.System.currentTimeMillis().toDouble()
    )
    return gps
  }

  private fun doCheckPermissions() {
    if (this.checkSelfPermission(Manifest.permission.READ_PHONE_STATE)
            != PackageManager.PERMISSION_GRANTED) {

      // Permission is not granted
      // Should we show an explanation?
      if (shouldShowRequestPermissionRationale(Manifest.permission.READ_PHONE_STATE)) {
        // Show an explanation to the user *asynchronously* -- don't block
        // this thread waiting for the user's response! After the user
        // sees the explanation, try again to request the permission.
      } else {
        // No explanation needed, we can request the permission.
        requestPermissions(arrayOf(Manifest.permission.READ_PHONE_STATE),
                APP_ON_PERMISSION_RESULT_READ_PHONE_STATE)
      }
    } else {
      // Permission has already been granted
    }
  }

  override fun onRequestPermissionsResult(requestCode: Int,
                                          permissions: Array<String>,
                                          grantResults: IntArray) {
    if (requestCode == APP_ON_PERMISSION_RESULT_READ_PHONE_STATE) {
      for(i in permissions.indices) {
        var permission = permissions[i]
        var grantResult = grantResults[i]
        if (permission.equals(Manifest.permission.READ_PHONE_STATE)) {
          if (grantResult == android.content.pm.PackageManager.PERMISSION_GRANTED) {
            // Can ask again...
          } else {
            requestPermissions(arrayOf(Manifest.permission.READ_PHONE_STATE), APP_ON_PERMISSION_RESULT_READ_PHONE_STATE)
          }
        }
      }
    }
  }


}
