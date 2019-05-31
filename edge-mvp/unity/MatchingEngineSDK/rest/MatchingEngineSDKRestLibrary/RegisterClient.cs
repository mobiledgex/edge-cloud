using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class RegisterClientRequest
  {
    [DataMember]
    public UInt32 ver;
    [DataMember]
    public string dev_name;
    [DataMember]
    public string app_name;
    [DataMember]
    public string app_vers;
    [DataMember]
    public string carrier_name;
    [DataMember]
    public string auth_token;
  }

  [DataContract]
  public class RegisterClientReply
  {
    [DataMember]
    public UInt32 Ver;

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

    [DataMember]
    public string session_cookie;
    [DataMember]
    public string token_server_uri;
  }

}
