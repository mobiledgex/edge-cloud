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
    [DataMember]
    public string Status = DistributedMatchEngine.ReplyStatus.RS_UNDEFINED.ToString();
    [DataMember]
    public string SessionCookie;
    [DataMember]
    public string TokenServerURI;
  }

}
