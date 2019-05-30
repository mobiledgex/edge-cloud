using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class GetLocationRequest
  {
    [DataMember]
    public UInt32 ver;
    [DataMember]
    public string session_cookie;
    [DataMember]
    public string carrier_name;
  }

  [DataContract]
  public class GetLocationReply
  {
    public enum LocStatus
    {
      LOC_UNKNOWN = 0,
      LOC_FOUND = 1,
      // The user does not allow his location to be tracked
      LOC_DENIED = 2
    }
    [DataMember]
    public UInt32 ver;

    public LocStatus status = LocStatus.LOC_UNKNOWN;

    [DataMember(Name = "status")]
    private string loc_status_string
    {
      get
      {
        return status.ToString();
      }
      set
      {
        status = Enum.TryParse(value, out LocStatus locStatus) ? locStatus : LocStatus.LOC_UNKNOWN;
      }
    }

    [DataMember]
    public string carrier_name;
    [DataMember]
    public string tower;
    [DataMember]
    public Loc network_location;
  }
}
