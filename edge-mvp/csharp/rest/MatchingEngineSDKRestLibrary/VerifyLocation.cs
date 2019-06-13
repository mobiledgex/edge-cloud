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
    public string session_cookie;
    [DataMember]
    public string carrier_name;
    [DataMember]
    public Loc gps_location;
    [DataMember]
    public string verify_loc_token;
  };

  [DataContract]
  public class VerifyLocationReply
  {
    // Status of the reply
    public enum TowerStatus
    {
      TOWER_UNKNOWN = 0,
      CONNECTED_TO_SPECIFIED_TOWER = 1,
      NOT_CONNECTED_TO_SPECIFIED_TOWER = 2,
    }

    public enum GPSLocationStatus
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

    public TowerStatus tower_status = TowerStatus.TOWER_UNKNOWN;

    [DataMember(Name = "tower_status")]
    private string tower_status_tring
    {
      get
      {
        return tower_status.ToString();
      }
      set
      {
        tower_status = Enum.TryParse(value, out TowerStatus towerStatus) ? towerStatus : TowerStatus.TOWER_UNKNOWN;
      }
    }

    public GPSLocationStatus gps_location_status = GPSLocationStatus.LOC_UNKNOWN;

    [DataMember(Name = "gps_location_status")]
    private string gps_location_status_string
    {
      get
      {
        return gps_location_status.ToString();
      }
      set
      {
        gps_location_status = Enum.TryParse(value, out GPSLocationStatus gpsLocation) ? gpsLocation : GPSLocationStatus.LOC_UNKNOWN;
      }
    }

    [DataMember]
    public double gps_location_accuracy_km;
  }
}
