using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class AppFqdn
  {
    // App  Name
    [DataMember]
    public string app_ame;
    // App Version
    [DataMember]
    public string app_vers;
    // developer name
    [DataMember]
    public string dev_name;
    // App FQDN
    [DataMember]
    public string[] fqdns;
    // optional android package name
    [DataMember]
    public string android_package_name;
  }

  [DataContract]
  public class FqdnListRequest
  {
    [DataMember]
    public UInt32 ver;
    [DataMember]
    public string session_cookie;
  };

  [DataContract]
  public class FqdnListReply
  {
    // Status of the reply
    public enum FLStatus
    {
      FL_UNDEFINED = 0,
      FL_SUCCESS = 1,
      FL_FAIL = 2
    }

    [DataMember]
    // API version
    public UInt32 ver;

    [DataMember]
    public AppFqdn[] app_fqdns;

    public FLStatus status = FLStatus.FL_UNDEFINED;

    [DataMember(Name = "status")]
    private string fl_status_string
    {
      get
      {
        return status.ToString();
      }
      set
      {
        status = Enum.TryParse(value, out FLStatus flStatus) ? flStatus : FLStatus.FL_UNDEFINED;
      }
    }
  }
}
