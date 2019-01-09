﻿using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public enum DlgCommType
  {
    DlgUndefined = 0,
    DlgSecure = 1,
    DlgOpen = 2
  }

  [Serializable]
  public class DynamicLocGroupRequest
  {
    public UInt32 Ver;
    // Session Cookie from RegisterClientRequest
    public string SessionCookie;
    public UInt64 LgId;
    public string CommType = DlgCommType.DlgUndefined.ToString();
    public string UserData;
  }

  [Serializable]
  public class DynamicLocGroupReply
  {
    
    public UInt32 Ver;
    // Status of the reply
    public string ReplyStatus = DistributedMatchEngine.ReplyStatus.RS_UNDEFINED.ToString();
    // Status of the reply
    public UInt32 ErrorCode;
    // Group Cookie for Secure Group Communication
    public string GroupCookie;
  }
}
