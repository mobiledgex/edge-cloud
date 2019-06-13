using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class Appinstance
  {
    // App Instance Name
    [DataMember]
    string app_name;
    // App Instance Version
    [DataMember]
    string app_vers;
    // App Instance FQDN
    [DataMember]
    string fqdn;
    // ports to access app
    [DataMember]
    AppPort[] ports;
  }

  [DataContract]
  public class CloudletLocation
  {
    // The carrier name that user is connected to ("Cellular Carrier Name")
    [DataMember]
    string carrier_name;
    // Cloudlet Name
    [DataMember]
    string cloudlet_name;
    // The GPS Location of the user
    Loc gps_location;
    [DataMember]
    // Distance of cloudlet vs loc in request
    double distance;
    // App instances
    [DataMember]
    Appinstance[] appinstances;
  }

  [DataContract]
  public class AppInstListRequest
  {
    [DataMember]
    public UInt32 ver;
    [DataMember]
    public string session_cookie;
    [DataMember]
    public string carrier_name;
    [DataMember]
    public Loc gps_location;
  }

  [DataContract]
  public class AppInstListReply
  {
    // Status of the reply
    public enum AIStatus
    {
      AI_UNDEFINED = 0,
      AI_SUCCESS = 1,
      AI_FAIL = 2
    }

    [DataMember]
    public UInt32 ver;

    public AIStatus status;

    [DataMember(Name = "status")]
    private string ai_status_string
    {
      get
      {
        return status.ToString();
      }
      set
      {
        status = Enum.TryParse(value, out AIStatus rStatus) ? rStatus : AIStatus.AI_UNDEFINED;
      }
    }

    [DataMember]
    public CloudletLocation[] cloudlets;
  }
}
