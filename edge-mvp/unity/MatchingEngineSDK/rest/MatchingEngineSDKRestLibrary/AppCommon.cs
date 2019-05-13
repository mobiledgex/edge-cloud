using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public enum LProto
  {
    // Unknown protocol
    LProtoUnknown = 0,
    // TCP (L4) protocol
    LProtoTCP = 1,
    // UDP (L4) protocol
    LProtoUDP = 2,
    // HTTP (L7 tcp) protocol
    LProtoHTTP = 3
  }

  [DataContract]
  public class AppPort
  {
    // TCP (L4), UDP (L4), or HTTP (L7) protocol
    public LProto proto = LProto.LProtoUnknown;

    [DataMember(Name = "proto")]
    private string protoString
    {
      get
      {
        return proto.ToString();
      }
      set
      {
        proto = Enum.TryParse(value, out LProto lproto) ? lproto : LProto.LProtoUnknown;
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
    public string FQDN_prefix;
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
    public Int64 seconds;
    public Int32 nanos;
  }
}
