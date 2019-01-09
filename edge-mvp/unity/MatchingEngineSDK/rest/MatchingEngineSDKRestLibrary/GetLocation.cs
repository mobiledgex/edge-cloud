﻿using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [Serializable]
  public class GetLocationRequest
  {
    public UInt32 Ver;
    public string SessionCookie;
    public string CarrierName;
  }

  [Serializable]
  public class GetLocationReply
  {
    public enum Loc_Status
    {
      LOC_UNKNOWN = 0,
      LOC_FOUND = 1,
      // The user does not allow his location to be tracked
      LOC_DENIED = 2
    }
    
    public UInt32 Ver;
    public string Status = Loc_Status.LOC_UNKNOWN.ToString();
    public string CarrierName;
    public string Tower;
    public Loc NetworkLocation;
  }
}
