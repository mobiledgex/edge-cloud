using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class GetLocationRequest
  {
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string CarrierName;
  }

  [DataContract]
  public class GetLocationReply
  {
    public enum Loc_Status
    {
      LOC_UNKNOWN = 0,
      LOC_FOUND = 1,
      // The user does not allow his location to be tracked
      LOC_DENIED = 2
    }
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string Status = Loc_Status.LOC_UNKNOWN.ToString();
    [DataMember]
    public string CarrierName;
    [DataMember]
    public string Tower;
    [DataMember]
    public Loc NetworkLocation;
  }
}
