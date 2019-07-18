using System;
using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class QosPosition
  {
    [DataMember]
    public string positionid; // *NOT* UInt64 for the purposes of REST.
    [DataMember]
    public Loc gps_location;
  }

  [DataContract]
  public class QosPositionKpiRequest
  {
    // API version
    [DataMember]
    public UInt32 ver;
    // Session Cookie from RegisterClientRequest
    [DataMember]
    public string session_cookie;
    // list of positions
    [DataMember]
    public QosPosition[] positions;
  }

  [DataContract]
  public class QosPositionResult
  {
    // as set by the client, must be unique within one QosPositionKpiRequest
    [DataMember]
    public Int64 positionid;
    // the location which was requested
    [DataMember]
    public Loc gps_location;
    // throughput 
    [DataMember]
    public float dluserthroughput_min;
    [DataMember]
    public float dluserthroughput_avg;
    [DataMember]
    public float dluserthroughput_max;
    [DataMember]
    public float uluserthroughput_min;
    [DataMember]
    public float uluserthroughput_avg;
    [DataMember]
    public float uluserthroughput_max;
    [DataMember]
    public float latency_min;
    [DataMember]
    public float latency_avg;
    [DataMember]
    public float latency_max;
  }

  [DataContract]
  public class QosPositionKpiReply
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

    // kpi details
    [DataMember]
    public QosPositionResult[] position_results;
  }

  [DataContract]
  public class QosPositionKpiStreamReply
  {
    [DataMember]
    public QosPositionKpiReply result;
    [DataMember]
    public RuntimeStreamError error;
  }

}
