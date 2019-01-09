﻿using System;

namespace DistributedMatchEngine
{
  [Serializable]
  public class RegisterClientRequest
  {
    public UInt32 Ver;
    public string DevName;
    public string AppName;
    public string AppVers;
    public string CarrierName;
    public string AuthToken;
  }

  [Serializable]
  public class RegisterClientReply
  {
    public UInt32 Ver;
    public string ReplyStatus = DistributedMatchEngine.ReplyStatus.RS_UNDEFINED.ToString();
    public string SessionCookie;
    public string TokenServerURI;
  }

}
