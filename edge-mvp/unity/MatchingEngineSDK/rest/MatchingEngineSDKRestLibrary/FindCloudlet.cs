﻿using System;

namespace DistributedMatchEngine
{
  [Serializable]
  public class FindCloudletRequest
  {
    public UInt32 Ver = 1;
    public string SessionCookie;
    public string CarrierName;
    public Loc GpsLocation;
    public string DevName;
    public string AppName;
    public string AppVers;
  }

  [Serializable]
  public class FindCloudletReply
  {
    // Standard Enum. Serializable Enum is converted to int64, not string.
    public enum FindStatus
    {
      FIND_UNKNOWN = 0,
      FIND_FOUND = 1,
      FIND_NOTFOUND = 2
    }

    
    public UInt32 Ver;
    public string status = FindStatus.FIND_UNKNOWN.ToString();
    public string FQDN;
    public AppPort[] ports;
    public Loc cloudlet_location;
  }

}
