using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class VerifyLocationRequest
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
    public string VerifyLocToken;
  };

  [DataContract]
  public class VerifyLocationReply
  {
    // Status of the reply
    public enum Tower_Status
    {
      TOWER_UNKNOWN = 0,
      CONNECTED_TO_SPECIFIED_TOWER = 1,
      NOT_CONNECTED_TO_SPECIFIED_TOWER = 2,
    }

    public enum GPS_Location_Status
    {
      LOC_UNKNOWN = 0,
      LOC_VERIFIED = 1,
      LOC_MISMATCH_SAME_COUNTRY = 2,
      LOC_MISMATCH_OTHER_COUNTRY = 3,
      LOC_ROAMING_COUNTRY_MATCH = 4,
      LOC_ROAMING_COUNTRY_MISMATCH = 5,
      LOC_ERROR_UNAUTHORIZED = 6,
      LOC_ERROR_OTHER = 7
    }

    [DataMember]
    public UInt32 ver;
    [DataMember]
    public string tower_status = Tower_Status.TOWER_UNKNOWN.ToString();
    [DataMember]
    public string gps_location_status = GPS_Location_Status.LOC_UNKNOWN.ToString();
    [DataMember]
    public double GPS_Location_Accuracy_KM;
  }
}
