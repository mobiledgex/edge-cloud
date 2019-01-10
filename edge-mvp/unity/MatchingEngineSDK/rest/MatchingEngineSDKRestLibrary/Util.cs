using System;
using System.IO;
using System.Runtime.Serialization;
using System.Threading.Tasks;

namespace DistributedMatchEngine
{
  public class Util
  {
    public Util()
    {

    }

    /*
     * Stream is consumed after use, not always repositionable.
     */
    public static string StreamToString(Stream ms)
    {
      ms.Position = 0;
      StreamReader reader = new StreamReader(ms);
      string jsonStr = reader.ReadToEnd();
      return jsonStr;
    }

    public async static Task<Loc> GetLocationFromDevice()
    {
      // FIXME: Do async device location.
      long timeLongMs = new DateTimeOffset(DateTime.UtcNow).ToUnixTimeMilliseconds();
      long seconds = timeLongMs / 1000;
      int nanoSec = (int)(timeLongMs % 1000) * 1000000;
      var ts = new Timestamp { nanos = nanoSec, seconds = seconds };
      var loc = new Loc()
      {
        course = 0,
        altitude = 100,
        horizontal_accuracy = 5,
        speed = 2,
        longitude = 37.459601,
        latitude = -122.149349,
        vertical_accuracy = 20,
        timestamp = ts
      };

      var dummyResult = await Task.FromResult(0);
      return loc;
    }
  }
}
