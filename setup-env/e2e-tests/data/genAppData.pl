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
my $GenplatosApp = 0;
my $NoAutoCluster = 0;
my $Nox1medium = 0; # without x1.medium, none of the prometheus apps can be created
############################################################################

GetOptions(
     'debug' => \$Debug,
     'app=s' => \$App,
     'operator=s' => \$Operator,
     'max=i' =>\$Max,
     'developer=s' => \$Developer,
     'cloudlet=s' => \$Cloudlet,
     'gencloudletinfo' => \$GenCloudletInfo,
     'genplatformapp' => \$GenplatosApp,
     'noautocluster' => \$NoAutoCluster,
     'nox1medium' => \$Nox1medium,
     'help' => \$Help
          ) or die "Invalid options \n $Usage";

# genplatformapp is platos enabling layer app
# if this option is passed we will create platos enabling layer app and a developer for it
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
  my $notifysrvport = 51000 + $cid;
  my $flavor = "x1.medium";
  if ($Nox1medium) {
    $flavor = "x1.tiny";
  }
  print ("
- key:
    organization: $operator
    name: $operator-$cloudlet-$cid
  accessuri: $operator-$cloudlet.$cid
  location:
    latitude: $lat
    longitude: $long
  ipsupport: IpSupportDynamic
  numdynamicips: 254
  platformtype: PlatformTypeFake
  notifysrvaddr: 127.0.0.1:$notifysrvport
  flavor:
    name: $flavor
\n")
}

sub printCloudletInfo{
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  print ("
- key:
    organization: $operator
    name: $operator-$cloudlet-$cid

  state: CloudletStateReady
  osmaxram: 65536
  osmaxvcores: 16
  osmaxvolgb: 500
\n")
}

sub printClusterInst{
  # only for no autocluster
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  my $lat = shift;
  my $long = shift;
  my $app = shift;

  print ("
- key:
    clusterkey:
      name: $app
    cloudletkey:
      organization: $operator
      name: $operator-$cloudlet-$cid
    organization: $Developer
  flavor:
    name: x1.small
  deployment: kubernetes
  nummasters: 1
  numnodes: 1
");
}

sub printAppinst{
  my $operator = shift;
  my $cloudlet = shift;
  my $cid = shift;
  my $lat = shift;
  my $long = shift;
  my $app = shift;

  my $clustername = "autocluster$app";
  my $clusterorg = "MobiledgeX";
  if ($NoAutoCluster) {
    $clustername = "$app";
    $clusterorg = "$Developer";
  }

  print ("
- key:
    appkey:
      organization: $Developer
      name: $app
      version: \"1.0\"
    clusterinstkey:
      clusterkey:
        name: $clustername
      cloudletkey:
        organization: $operator
        name: $operator-$cloudlet-$cid
      organization: $clusterorg
  cloudletloc:
    latitude: $lat
    longitude: $long
");

}

sub genLatLongs{
  my $type = shift;
  my $c = 0;

  # if multiple apps are specified with the same name, they
  # target the same clusterinst.
  my %clusterinsts = ();
  
  print "$type:";
  for (my $lat = 1;$lat <= $MAXLAT;$lat++){
    for (my $long = 1;$long <= $MAXLONG;$long++){
      $c++;
      if ($c > $Max){
	return;
      }
      my $operator = $Operator;
      # the last 2 will be azure and gcp if we have 5
      if ($Max >= 5) {
        if ($c == $Max-1) {
          $operator = "azure";
        }
        if ($c == $Max){
          $operator = "gcp";
        }
      }
      if ($type eq "cloudlets"){
        printCloudlet($operator,$Cloudlet,$c,$lat,$long);
      }
      if ($type eq "cloudletinfos"){
        printCloudletInfo($operator,$Cloudlet,$c)
      }
      if ($type eq "clusterinsts"){
        my @apps = split(",", $App);
        foreach my $app(@apps){
          if ($clusterinsts{"$operator$Cloudlet$c$app"}) {
            next;
          }
          printClusterInst($operator,$Cloudlet,$c,$lat,$long,$app);
          $clusterinsts{"$operator$Cloudlet$c$app"} = 1;
        }
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
operatorcodes:
- code: wifi
  organization: $Operator
")
}


sub genplatosDeveloper{
  print("
- key:
    name: platos
  username: platos-user
  passhash: 123456789012345670
  address: 1234 platos street
  email: platos\@gmail.com
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
  disk: 2
");
 if ($Nox1medium) {
   print("\n");
   return;
 }
 print("- key:
    name: x1.medium
  ram: 4096
  vcpus: 4
  disk: 4\n\n");
}


sub genApp{
  my $app = shift;
  my $androidpackagename = lc("com.$Developer.$app");
  my $fqdn = lc("$app.$Developer.com");
  print(
"- key:
    organization: $Developer
    name: $app
    version : \"1.0\"
  imagetype: ImageTypeDocker
  defaultflavor:
    name: x1.small
  accessports: tcp:80,tcp:443,udp:10002
  officialfqdn: $fqdn");

  # if this is a platrfom app we need to add android package
  if ($GenplatosApp) {
    print("
  androidpackagename: $androidpackagename");
  }
  print("
  authpublickey: \"-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Spdynjh+MPcziCH2Gij\\nTkK9fspTH4onMtPTgxo+MQC+OZTwetvYFJjGV8jnYebtuvWWUCctYmt0SIPmA0F0\\nVU6qzSlrBOKZ9yA7Rj3jSQtNrI5vfBIzK1wPDm7zuy5hytzauFupyfboXf4qS4uC\\nGJCm9EOzUSCLRryyh7kTxa4cYHhhTTKNTTy06lc7YyxBsRsN/4jgxjjkxe3J0SfS\\nz3eaHmfFn/GNwIAqy1dddTJSPugRkK7ZjFR+9+sscY9u1+F5QPwxa8vTB0U6hh1m\\nQnhVd1d9osRwbyALfBY8R+gMgGgEBCPYpL3u5iSjgD6+n4d9RQS5zYRpeMJ1fX0C\\n/QIDAQAB\\n-----END PUBLIC KEY-----\\n\"
")
}



sub genplatosApp{
print(
"- key:
    organization: platos
    name: PlatosEnablingLayer
    version: \"1.0\"
  imagepath: registry.mobiledgex.net/dummyvalue
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
  $extra = $extra . " -gencloudletinfo";
}
if ($GenplatosApp) {
  $extra = $extra . " -genplatformapp";
}
if ($Cloudlet ne "cloud") {
  $extra = $extra . " -cloudlet " . $Cloudlet;
}
if ($NoAutoCluster) {
  $extra = $extra . " -noautocluster";
}
if ($Nox1medium) {
  $extra = $extra . " -nox1medium";
}
print ("# Generated by genAppData.pl as follows:
#   ./genAppData.pl -operator $Operator -developer $Developer -app $App -max $Max $extra\n");

genFlavor();
genOperator();
genLatLongs("cloudlets");
if ($GenCloudletInfo) {
  genLatLongs("cloudletinfos");
}
if ($NoAutoCluster) {
  genLatLongs("clusterinsts");
} else {
  # trigger delete of reservable auto clusters (only for delete)
  print("idlereservableclusterinsts:
  idletime: 0
");
}
print "\napps:\n";
if ($GenplatosApp) {
  genplatosApp();
}
# if multiple apps are specified with the same name, don't
# generate multiple App defs
my %appnames = ();
my @apps = split(",", $App);
foreach my $app(@apps){
  if ($appnames{$app}) {
    next;
  }
  genApp($app);
  $appnames{$app} = 1;
}
genLatLongs("appinstances");

