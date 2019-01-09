﻿using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [Serializable]
  public class AppFqdn
  {
    // App  Name
    public string AppName;
    // App Version
    public string AppVers;
    // developer name
    public string DevName;
    // App FQDN
    public string FQDN;
    // optional android package name
    public string AndroidPackageName;
  }

  [Serializable]
  public class FqdnListRequest
  {
    public UInt32 Ver;
    public string SessionCookie;
  };

  [Serializable]
  public class FqdnListReply
  {
    // Status of the reply
    public enum FL_Status
    {
      FL_UNDEFINED = 0,
      FL_SUCCESS = 1,
      FL_FAIL = 2
    }
    
    public AppFqdn[] AppFqdns;
    public string Status = FL_Status.FL_UNDEFINED.ToString();
  }
}
