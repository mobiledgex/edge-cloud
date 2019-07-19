using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public enum LProto
  {
    // Unknown protocol
    L_PROTO_UNKNOWN = 0,
    // TCP (L4) protocol
    L_PROTO_TCP = 1,
    // UDP (L4) protocol
    L_PROTO_UDP = 2,
    // HTTP (L7 tcp) protocol
    L_PROTO_HTTP = 3
  }

  [DataContract]
  public class AppPort
  {
    // TCP (L4), UDP (L4), or HTTP (L7) protocol
    public LProto proto = LProto.L_PROTO_UNKNOWN;

    [DataMember(Name = "proto")]
    private string proto_string
    {
      get
      {
        return proto.ToString();
      }
      set
      {
        proto = Enum.TryParse(value, out LProto lproto) ? lproto : LProto.L_PROTO_UNKNOWN;
      }
    }

    // Container port
    [DataMember]
    public Int32 internal_port;
    // Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)
    [DataMember]
    public Int32 public_port;
    // Public facing path for HTTP L7 access.
    [DataMember]
    public string path_prefix;
    // FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.
    [DataMember]
    public string fqdn_prefix;
  }

  public enum IDTypes
  {
    ID_UNDEFINED = 0,
    IMEI = 1,
    MSISDN = 2,
    IPADDR = 3
  }

  public enum ReplyStatus
  {
    RS_UNDEFINED = 0,
    RS_SUCCESS = 1,
    RS_FAIL = 2
  }

  public class Timestamp
  {
    public string seconds;
    public Int32 nanos;
  }
}
