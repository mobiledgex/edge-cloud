using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class RegisterClientRequest
  {
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string DevName;
    [DataMember]
    public string AppName;
    [DataMember]
    public string AppVers;
    [DataMember]
    public string CarrierName;
    [DataMember]
    public string AuthToken;
  }

  [DataContract]
  public class RegisterClientReply
  {
    [DataMember]
    public UInt32 Ver;

    public ReplyStatus Status = ReplyStatus.RS_UNDEFINED;

    [DataMember(Name = "Status")]
    private string ReplyStatusString
    {
      get
      {
        return Status.ToString();
      }
      set
      {
        Status = Enum.TryParse(value, out ReplyStatus replyStatus) ? replyStatus : ReplyStatus.RS_UNDEFINED;
      }
    }

    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string TokenServerURI;
  }

}
