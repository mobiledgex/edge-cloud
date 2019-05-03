using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class AppFqdn
  {
    // App  Name
    [DataMember]
    public string AppName;
    // App Version
    [DataMember]
    public string AppVers;
    // developer name
    [DataMember]
    public string DevName;
    // App FQDN
    [DataMember]
    public string FQDN;
    // optional android package name
    [DataMember]
    public string AndroidPackageName;
  }

  [DataContract]
  public class FqdnListRequest
  {
    [DataMember]
    public UInt32 Ver;
    [DataMember]
    public string SessionCookie;
  };

  [DataContract]
  public class FqdnListReply
  {
    // Status of the reply
    public enum FL_Status
    {
      FL_UNDEFINED = 0,
      FL_SUCCESS = 1,
      FL_FAIL = 2
    }
    [DataMember]
    public AppFqdn[] AppFqdns;

    public FL_Status Status = FL_Status.FL_UNDEFINED;

    [DataMember(Name = "Status")]
    private string FlStatusString
    {
      get
      {
        return Status.ToString();
      }
      set
      {
        Status = Enum.TryParse(value, out FL_Status flStatus) ? flStatus : FL_Status.FL_UNDEFINED;
      }
    }
  }
}
