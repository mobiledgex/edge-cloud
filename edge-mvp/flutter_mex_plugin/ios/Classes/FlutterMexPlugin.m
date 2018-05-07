#import "FlutterMexPlugin.h"
#import <flutter_mex_plugin/flutter_mex_plugin-Swift.h>

@implementation FlutterMexPlugin
+ (void)registerWithRegistrar:(NSObject<FlutterPluginRegistrar>*)registrar {
  [SwiftFlutterMexPlugin registerWithRegistrar:registrar];
}
@end
