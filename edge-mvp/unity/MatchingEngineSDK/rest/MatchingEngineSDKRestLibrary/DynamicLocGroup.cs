using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public enum DlgCommType
  {
    DLG_UNDEFINED = 0,
    DLG_SECURE = 1,
    DLG_OPEN = 2
  }

  [DataContract]
  public class DynamicLocGroupRequest
  {
    [DataMember]
    public UInt32 ver;
    // Session Cookie from RegisterClientRequest
    [DataMember]
    public string session_cookie;
    [DataMember]
    public UInt64 lg_id;

    public DlgCommType comm_type = DlgCommType.DLG_UNDEFINED;

    [DataMember(Name = "comm_type")]
    private string comm_type_string
    {
      get
      {
        return comm_type.ToString();
      }
      set
      {
        comm_type = Enum.TryParse(value, out DlgCommType commType) ? commType : DlgCommType.DLG_UNDEFINED;
      }
    }

    [DataMember]
    public string user_data;
  }

  [DataContract]
  public class DynamicLocGroupReply
  {
    [DataMember]
    public UInt32 ver;

    // Status of the reply
    public ReplyStatus status = ReplyStatus.RS_UNDEFINED;

    [DataMember(Name = "status")]
    private string reply_status_string
    {
      get
      {
        return status.ToString();
      }
      set
      {
        status = Enum.TryParse(value, out ReplyStatus replyStatus) ? replyStatus : ReplyStatus.RS_UNDEFINED;
      }
    }

    // Error Code based on Failure
    [DataMember]
    public UInt32 error_code;
    // Group Cookie for Secure Group Communication
    [DataMember]
    public string group_cookie;
  }
}
