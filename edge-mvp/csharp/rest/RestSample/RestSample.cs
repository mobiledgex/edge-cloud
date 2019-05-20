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

        // Start location task. This is for test use only. The source of the
        // location in an Unity application should be from an application context
        // LocationService.
        var locTask = Util.GetLocationFromDevice();

        var registerClientRequest = me.CreateRegisterClientRequest(carrierName, devName, appName, appVers, developerAuthToken);

        // APIs depend on Register client to complete successfully:
        try
        {
          var registerClientReply = await me.RegisterClient(host, port, registerClientRequest);
          Console.WriteLine("RegisterClient Reply Status: " + registerClientReply.Status);
        }
        catch (HttpException httpe) // HTTP status, and REST API call error codes.
        {
          // server error code, and human readable message:
          Console.WriteLine("RegisterClient Exception: " + httpe.Message + ", HTTP StatusCode: " + httpe.HttpStatusCode + ", API ErrorCode: " + httpe.ErrorCode + "\nStack: " + httpe.StackTrace);
        }
        // Do Verify and FindCloudlet in concurrent tasks:
        var loc = await locTask;

        // Independent requests:
        var verifyLocationRequest = me.CreateVerifyLocationRequest(carrierName, loc);
        var findCloudletRequest = me.CreateFindCloudletRequest(carrierName, devName, appName, appVers, loc);
        var getLocationRequest = me.CreateGetLocationRequest(carrierName);


        // These are asynchronious calls, of independent REST APIs.

        // FindCloudlet:
        try
        {
          var findCloudletReply = await me.FindCloudlet(host, port, findCloudletRequest);
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
                  ", path_prefix: " + p.path_prefix);
          }
        }
        catch (HttpException httpe)
        {
          Console.WriteLine("FindCloudlet Exception: " + httpe.Message + ", HTTP StatusCode: " + httpe.HttpStatusCode + ", API ErrorCode: " + httpe.ErrorCode + "\nStack: " + httpe.StackTrace);
        }

        // Get Location:
        try
        {
          var getLocationReply = await me.GetLocation(host, port, getLocationRequest);
          var location = getLocationReply.NetworkLocation;
          Console.WriteLine("GetLocationReply: longitude: " + location.longitude + ", latitude: " + location.latitude);
        }
        catch (HttpException httpe)
        {
          Console.WriteLine("GetLocation Exception: " + httpe.Message + ", HTTP StatusCode: " + httpe.HttpStatusCode + ", API ErrorCode: " + httpe.ErrorCode + "\nStack: " + httpe.StackTrace);
        }

        // Verify Location:
        try
        {
          Console.WriteLine("VerifyLocation() may timeout, due to reachability of carrier verification servers from your network.");
          var verifyLocationReply = await me.VerifyLocation(host, port, verifyLocationRequest);
          Console.WriteLine("VerifyLocation Reply: " + verifyLocationReply.gps_location_status);
        }
        catch (HttpException httpe)
        {
          Console.WriteLine("VerifyLocation Exception: " + httpe.Message + ", HTTP StatusCode: " + httpe.HttpStatusCode + ", API ErrorCode: " + httpe.ErrorCode + "\nStack: " + httpe.StackTrace);
        }
        catch (InvalidTokenServerTokenException itste)
        {
          Console.WriteLine(itste.Message + "\n" + itste.StackTrace);
        }
      }
      catch (Exception e) // Catch All
      {
        Console.WriteLine(e.Message + "\n" + e.StackTrace);
      }

    }
  };
}
