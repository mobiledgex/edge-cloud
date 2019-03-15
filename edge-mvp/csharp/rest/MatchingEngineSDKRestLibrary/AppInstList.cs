using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class Appinstance
  {
    // App Instance Name
    [DataMember]
    string AppName;
    // App Instance Version
    [DataMember]
    string AppVers;
    // App Instance FQDN
    [DataMember]
    string FQDN;
    // ports to access app
    [DataMember]
    AppPort[] ports;
  }

  [DataContract]
  public class CloudletLocation
  {
    // The carrier name that user is connected to ("Cellular Carrier Name")
    [DataMember]
    string CarrierName;
    // Cloudlet Name
    [DataMember]
    string CloudletName;
    // The GPS Location of the user
    Loc GpsLocation;
    [DataMember]
    // Distance of cloudlet vs loc in request
    double Distance;
    // App instances
    [DataMember]
    Appinstance[] Appinstances;
  }

  [DataContract]
  public class AppInstListRequest
  {
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string CarrierName;
    [DataMember]
    public Loc GpsLocation;
  }

  [DataContract]
  public class AppInstListReply
  {
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string ReplyStatus;
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string TokenServerURI;
  }
}
