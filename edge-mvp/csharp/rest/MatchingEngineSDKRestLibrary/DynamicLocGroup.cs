using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public enum DlgCommType
  {
    DlgUndefined = 0,
    DlgSecure = 1,
    DlgOpen = 2
  }

  [DataContract]
  public class DynamicLocGroupRequest
  {
    [DataMember]
    public UInt32 Ver;
    // Session Cookie from RegisterClientRequest
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public UInt64 LgId;
    [DataMember]
    public string CommType = DlgCommType.DlgUndefined.ToString();
    [DataMember]
    public string UserData;
  }

  [DataContract]
  public class DynamicLocGroupReply
  {
    [DataMember]
    public UInt32 Ver;
    // Status of the reply
    [DataMember]
    public string ReplyStatus = DistributedMatchEngine.ReplyStatus.RS_UNDEFINED.ToString();
    // Status of the reply
    [DataMember]
    public UInt32 ErrorCode;
    // Group Cookie for Secure Group Communication
    [DataMember]
    public string GroupCookie;
  }
}
