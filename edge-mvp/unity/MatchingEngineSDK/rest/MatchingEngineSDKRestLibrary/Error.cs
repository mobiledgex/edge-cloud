using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  public class ProtobufAny
  {
    [DataMember]
    public string type_url;
    [DataMember]
    public string value; // byte
  }

  public class RuntimeStreamError
  {
    [DataMember]
    public Int32 grpc_code;
    [DataMember]
    Int32 http_code;
    [DataMember]
    public string message;
    [DataMember]
    public string http_status;
    [DataMember]
    public ProtobufAny details;
  }
}
