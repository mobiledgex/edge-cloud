#!/usr/bin/perl

use strict;
use Getopt::Long;



##########################################################################
##
##options
my $Debug = 0;
my $Usage = "Usage...\n";
my $App = 0;
my $Operator = 0;
my $Developer = 0;
my $Cloudlet = "cloud";
my $Max = 0;
my $GenCloudletInfo = 0;
my $Help = 0;
my $GenSamsungApp = 0;
############################################################################

GetOptions(
     'debug' => \$Debug,
     'app=s' => \$App,
     'operator=s' => \$Operator,
     'max=i' =>\$Max,
     'developer=s' => \$Developer,
     'cloudlet=s' => \$Cloudlet,
     'gencloudletinfo' => \$GenCloudletInfo,
     'genplatformapp' => \$GenSamsungApp,
     'help' => \$Help
          ) or die "Invalid options \n $Usage";

# genplatformapp is Samsung enabling layer app
# if this option is passed we will create Samsung enabling layer app and a developer for it
my $MAXLAT = 85;
my $MAXLONG = 175;

my $Usage = "
     genAppData.pl <options>
       options:
           -operator <Operator name>
           -app <App names list separated by comma, e.g. app1,app2>
           -developer <developer name>
           -cloudlet <cloudlet name>
           -genplatformapp
           -max <max cloudlets to create>
           -help <show this message>

     Example:
        ./genAppData.pl -operator OP1 -developer Dev1 -app app1,app2 -genplatformapp -max 100 > appdata_100.yml\n";



sub debug{
   my $string = shift;
   if ($Debug){
      print STDOUT "DEBUG: $string\n";
   }
}

sub checkOptions{
   my $rc = 1;

   if (!$Operator){
      $rc = 0;
      print "-operator required\n";
   }
   if (!$Developer){
      $rc = 0;
      print "-developer required\n";
   }
   if (!$Max){
      $rc = 0;
      print "-max required\n";
   }

   return $rc;
}#checkOptions

sub printCloudlet{
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  my $lat = shift;
  my $long = shift;
  print ("
- key:
    operatorkey:
      name: $operator
    name: $operator-$cloudlet-$cid
  accessuri: $operator-$cloudlet.$cid

  location:
    latitude: $lat
    longitude: $long
  ipsupport: IpSupportDynamic
  numdynamicips: 254
\n")
}

sub printCloudletInfo{
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  print ("
- key:
    operatorkey:
      name: $operator
    name: $operator-$cloudlet-$cid

  state: CloudletStateReady
  osmaxram: 65536
  osmaxvcores: 16
  osmaxvolgb: 500
\n")
}

sub printAppinst{
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  my $lat = shift;
  my $long = shift;
  my $app = shift;

  print ("
- key:
    appkey:
      developerkey:
        name: $Developer
      name: $app
      version: \"1.0\"
    clusterinstkey:
      clusterkey:
        name: autocluster
      cloudletkey:
        operatorkey:
          name: $operator
        name: $operator-$cloudlet-$cid
      developer: $Developer
  cloudletloc:
    latitude: $lat
    longitude: $long
");

}

sub genLatLongs{
  my $type = shift;
  my $c = 0;

  print "$type:";
  for (my $lat = 1;$lat <= $MAXLAT;$lat++){
    for (my $long = 1;$long <= $MAXLONG;$long++){
      $c++;
      if ($c > $Max){
	return;
      }
      my $operator = $Operator;
      # the last 2 will be azure and gcp if we have 10
      if ($c == 9) {
          $operator = "azure";
      }
      if ($c == 10){
          $operator = "gcp";
      }
      if ($type eq "cloudlets"){
        printCloudlet($operator,$Cloudlet,$c,$lat,$long);
      }
      if ($type eq "cloudletinfos"){
        printCloudletInfo($operator,$Cloudlet,$c)
      }
      if ($type eq "appinstances"){
        my @apps = split(",", $App);
        foreach my $app(@apps){
          printAppinst($operator,$Cloudlet,$c,$lat,$long,$app);
        }
      }
    }
  }
}



sub genOperator{
  print("
operators:
- key:
    name: $Operator
- key:
    name: azure
- key:
    name: gcp
")
}

sub genDeveloper{
  print("
developers:
- key:
    name: $Developer
  username: $Developer-user
  passhash: 123456789012345670
  address: 1234 $Developer street
  email: $Developer\@gmail.com
\n")
}

sub genSamsungDeveloper{
  print("
- key:
    name: Samsung
  username: Samsung-user
  passhash: 123456789012345670
  address: 1234 Samsung street
  email: samsung\@gmail.com
\n")
}

sub genFlavor{
 print("
flavors:
- key:
    name: x1.tiny
  ram: 1024
  vcpus: 1
  disk: 1
- key:
    name: x1.small
  ram: 2048
  vcpus: 2
  disk: 2\n\n")
}


sub genApp{
  my $app = shift;
  my $androidpackagename = lc("com.$Developer.$app");
  print(
"- key:
    developerkey:
       name: $Developer
    name: $app
    version : \"1.0\"
  imagetype: ImageTypeDocker
  defaultflavor:
    name: x1.small
  accessports: tcp:80,http:443,udp:10002");
  # if this is a platrfom app we need to add android package
  if ($GenSamsungApp) {
    print("
  androidpackagename: $androidpackagename
  permitsplatformapps: true");
  }
  print("
  authpublickey: \"-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Spdynjh+MPcziCH2Gij\\nTkK9fspTH4onMtPTgxo+MQC+OZTwetvYFJjGV8jnYebtuvWWUCctYmt0SIPmA0F0\\nVU6qzSlrBOKZ9yA7Rj3jSQtNrI5vfBIzK1wPDm7zuy5hytzauFupyfboXf4qS4uC\\nGJCm9EOzUSCLRryyh7kTxa4cYHhhTTKNTTy06lc7YyxBsRsN/4jgxjjkxe3J0SfS\\nz3eaHmfFn/GNwIAqy1dddTJSPugRkK7ZjFR+9+sscY9u1+F5QPwxa8vTB0U6hh1m\\nQnhVd1d9osRwbyALfBY8R+gMgGgEBCPYpL3u5iSjgD6+n4d9RQS5zYRpeMJ1fX0C\\n/QIDAQAB\\n-----END PUBLIC KEY-----\\n\"
")
}


sub genDefaultAppInst{
   my $app = shift;

   print(
"- key:
    appkey:
      developerkey:
        name: $Developer
      name: $app
      version: \"1.0\"
    clusterinstkey:
      clusterkey:
        name: default
      cloudletkey:
        operatorkey:
          name: developer
        name: default
  uri: default.$app.$Developer.com
\n")
}

sub genSamsungAppInst{
   print(
"- key:
    appkey:
      developerkey:
        name: Samsung
      name: SamsungEnablingLayer
      version: \"1.0\"
    clusterinstkey:
      clusterkey:
        name: default
      cloudletkey:
        operatorkey:
          name: developer
        name: default
  uri: default.samsungenablement.samsung.com
\n")
}

sub genSamsungApp{
print(
"- key:
    developerkey:
      name: Samsung
    name: SamsungEnablingLayer
    version: \"1.0\"
  imagepath: dummyvalue
  imagetype: ImageTypeDocker
  defaultflavor:
    name: x1.small
  accessports: \"tcp:64000\"
\n")
}

#main

if (!checkOptions()){
    print "ERROR: invalid options\n";
    print $Usage;
    exit 1;
}

if ($Help){
  print $Usage;
  exit 0;
}

my $extra = "";
if ($GenCloudletInfo) {
  $extra = $extra . " -gencloudletinfo"
}
if ($GenSamsungApp) {
  $extra = $extra . " -genplatformapp"
}
if ($Cloudlet ne "cloud") {
  $extra = $extra . " -cloudlet " . $Cloudlet
}
print ("# Generated by genAppData.pl as follows:
#   ./genAppData.pl -operator $Operator -developer $Developer -app $App -max $Max $extra\n");

genFlavor();
genOperator();
genLatLongs("cloudlets");
if ($GenCloudletInfo) {
  genLatLongs("cloudletinfos");
}
genDeveloper();
if ($GenSamsungApp) {
  genSamsungDeveloper();
}

print "apps:\n";
if ($GenSamsungApp) {
  genSamsungApp();
}
my @apps = split(",", $App);
foreach my $app(@apps){
   genApp($app);
}
genLatLongs("appinstances");

foreach my $app(@apps){
   genDefaultAppInst($app);
}
if ($GenSamsungApp) {
  genSamsungAppInst();
}
