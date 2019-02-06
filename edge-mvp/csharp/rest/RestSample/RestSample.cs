using System;
using System.Threading.Tasks;
using DistributedMatchEngine;

namespace RestSample
{
  class Program
  {
    static string carrierName = "tdg";
    static string appName = "EmptyMatchEngineApp";
    static string devName = "EmptyMatchEngineApp";
    static string appVers = "1.0";
    static string developerAuthToken = "";

    static string host = "tdg.dme.mobiledgex.net";
    static UInt32 port = 38001;

    // Get the ephemerial carriername from device specific properties.
    async static Task<string> getCurrentCarrierName()
    {
      var dummy = await Task.FromResult(0);
      return carrierName;
    }

    async static Task Main(string[] args)
    {
      try
      {
        carrierName = await getCurrentCarrierName();

        Console.WriteLine("RestSample!");

        MatchingEngine me = new MatchingEngine();
        port = MatchingEngine.defaultDmeRestPort;

        // Start location task:
        var locTask = Util.GetLocationFromDevice();

        var registerClientRequest = me.CreateRegisterClientRequest(carrierName, appName, devName, appVers, developerAuthToken);

        // Await synchronously.
        var registerClientReply = await me.RegisterClient(host, port, registerClientRequest);
        Console.WriteLine("Reply: Session Cookie: " + registerClientReply.SessionCookie);

        // Do Verify and FindCloudlet in parallel tasks:
        var loc = await locTask;

        var verifyLocationRequest = me.CreateVerifyLocationRequest(carrierName, loc);
        var findCloudletRequest = me.CreateFindCloudletRequest(carrierName, devName, appName, appVers, loc);
        var getLocationRequest = me.CreateGetLocationRequest(carrierName);


        // Async:
        var findCloudletTask = me.FindCloudlet(host, port, findCloudletRequest);
        var verfiyLocationTask = me.VerifyLocation(host, port, verifyLocationRequest);


        var getLocationTask = me.GetLocation(host, port, getLocationRequest);

        // Awaits:
        var findCloudletReply = await findCloudletTask;
        Console.WriteLine("FindCloudlet Reply: " + findCloudletReply.status);

        var verifyLocationReply = await verfiyLocationTask;
        Console.WriteLine("VerifyLocation Reply: " + verifyLocationReply.gps_location_status);

        var getLocationReply = await getLocationTask;
        var location = getLocationReply.NetworkLocation;
        Console.WriteLine("GetLocationReply: longitude: " + location.longitude + ", latitude: " + location.latitude);
      }
      catch (InvalidTokenServerTokenException itste)
      {
        Console.WriteLine(itste.StackTrace);
      }
      catch (Exception e)
      {
        Console.WriteLine(e.StackTrace);
      }

    }
  };
}
