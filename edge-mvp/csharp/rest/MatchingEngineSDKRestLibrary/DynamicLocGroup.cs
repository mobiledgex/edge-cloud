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

    public DlgCommType CommType = DlgCommType.DlgUndefined;

    [DataMember(Name = "CommType")]
    private string CommTypeString
    {
      get
      {
        return CommType.ToString();
      }
      set
      {
        CommType = Enum.TryParse(value, out DlgCommType commType) ? commType : DlgCommType.DlgUndefined;
      }
    }

    [DataMember]
    public string UserData;
  }

  [DataContract]
  public class DynamicLocGroupReply
  {
    [DataMember]
    public UInt32 Ver;
    // Status of the reply

    public ReplyStatus ReplyStatus = ReplyStatus.RS_UNDEFINED;

    [DataMember(Name = "Status")]
    private string ReplyStatusString
    {
      get
      {
        return ReplyStatus.ToString();
      }
      set
      {
        ReplyStatus = Enum.TryParse(value, out ReplyStatus replyStatus) ? replyStatus : ReplyStatus.RS_UNDEFINED;
      }
    }

    // Status of the reply
    [DataMember]
    public UInt32 ErrorCode;
    // Group Cookie for Secure Group Communication
    [DataMember]
    public string GroupCookie;
  }
}
