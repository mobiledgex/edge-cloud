﻿using System;

namespace DistributedMatchEngine
{
  [Serializable]
  public class Appinstance
  {
    // App Instance Name
    string AppName;
    // App Instance Version
    string AppVers;
    // App Instance FQDN
    string FQDN;
    // ports to access app
    AppPort[] ports;
  }

  [Serializable]
  public class CloudletLocation
  {
    // The carrier name that user is connected to ("Cellular Carrier Name")
    
    string CarrierName;
    // Cloudlet Name
    string CloudletName;
    // The GPS Location of the user
    Loc GpsLocation;
    // Distance of cloudlet vs loc in request
    double Distance;
    // App instances
    Appinstance[] Appinstances;
  }

  [Serializable]
  public class AppInstListRequest
  {
    public UInt32 Ver;
    public string SessionCookie;
    public string CarrierName;
    public Loc GpsLocation;
  }

  [Serializable]
  public class AppInstListReply
  {
    
    public UInt32 Ver;
    public string ReplyStatus;
    public string SessionCookie;
    public string TokenServerURI;
  }
}
