using System.Runtime.Serialization;

namespace DistributedMatchEngine
{
  [DataContract]
  public class Loc
  {
    [DataMember]
    public double latitude;
    [DataMember]
    public double longitude;
    [DataMember]
    public double horizontal_accuracy;
    [DataMember]
    public double vertical_accuracy;
    [DataMember]
    public double altitude;
    [DataMember]
    public double course;
    [DataMember]
    public double speed;
    [DataMember]
    public Timestamp timestamp;
  }
}
