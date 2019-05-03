using System;
using System.Threading.Tasks;
using DistributedMatchEngine;

namespace RestSample
{
  class Program
  {
    static string carrierName = "tdg";
    static string devName = "MobiledgeX";
    static string appName = "MobiledgeX SDK Demo";
    static string appVers = "1.0";
    static string developerAuthToken = "";

    static string host = "mexdemo.dme.mobiledgex.net";
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
        // GetLocationFromDevice is stubbed. It is NOT implemented yet, it requires GPS.
        var locTask = Util.GetLocationFromDevice();

        var registerClientRequest = me.CreateRegisterClientRequest(carrierName, devName, appName, appVers, developerAuthToken);

        // Await synchronously.
        var registerClientReply = await me.RegisterClient(host, port, registerClientRequest);
        Console.WriteLine("Reply: Session Cookie: " + registerClientReply.SessionCookie + ", Status: " + registerClientReply.Status);

        // Do Verify and FindCloudlet in concurrent tasks:
        var loc = await locTask;

        var verifyLocationRequest = me.CreateVerifyLocationRequest(carrierName, loc);
        var findCloudletRequest = me.CreateFindCloudletRequest(carrierName, devName, appName, appVers, loc);
        var getLocationRequest = me.CreateGetLocationRequest(carrierName);


        // Async:
        var findCloudletTask = me.FindCloudlet(host, port, findCloudletRequest);
        var getLocationTask = me.GetLocation(host, port, getLocationRequest);

        // Awaits:
        var findCloudletReply = await findCloudletTask;
        Console.WriteLine("FindCloudlet Reply: " + findCloudletReply.status);
        Console.WriteLine("FindCloudlet:" +
                " Ver: " + findCloudletReply.Ver +
                ", FQDN: " + findCloudletReply.FQDN +
                ", cloudlet_location: " +
                " long: " + findCloudletReply.cloudlet_location.longitude +
                ", lat: " + findCloudletReply.cloudlet_location.latitude);
        // App Ports:
        foreach (AppPort p in findCloudletReply.ports)
        {
          Console.WriteLine("Port: FQDN_prefix: " + p.FQDN_prefix +
                ", protocol: " + p.proto +
                ", public_port: " + p.public_port +
                ", internal_port: " + p.internal_port +
                ", public_path: " + p.public_path);
        }



        // A MobiledgeX enabled carrier is required for these two APIs:
        var getLocationReply = await getLocationTask;
        var location = getLocationReply.NetworkLocation;
        Console.WriteLine("GetLocationReply: longitude: " + location.longitude + ", latitude: " + location.latitude);

        Console.WriteLine("VerifyLocation() may timeout, due to reachability of carrier verification servers from your network.");
        var verifyLocationReply = await me.VerifyLocation(host, port, verifyLocationRequest);
        Console.WriteLine("VerifyLocation Reply: " + verifyLocationReply.gps_location_status);
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
