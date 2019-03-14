using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class FindCloudletRequest
  {
    [DataMember]
    public UInt32 Ver = 1;
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string CarrierName;
    [DataMember]
    public Loc GpsLocation;
    [DataMember]
    public string DevName;
    [DataMember]
    public string AppName;
    [DataMember]
    public string AppVers;
  }

  [DataContract]
  public class FindCloudletReply
  {
    // Standard Enum. DataContract Enum is converted to int64, not string.
    public enum FindStatus
    {
      FIND_UNKNOWN = 0,
      FIND_FOUND = 1,
      FIND_NOTFOUND = 2
    }

    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string status = FindStatus.FIND_UNKNOWN.ToString();
    [DataMember]
    public string FQDN;
    [DataMember]
    public AppPort[] ports;
    [DataMember]
    public Loc cloudlet_location;
  }

}
