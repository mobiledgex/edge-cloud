using System;

using Grpc.Core;
using System.Net;
using System.Diagnostics;

// MobiledgeX Matching Engine API.
using DistributedMatchEngine;

namespace MexGrpcSampleConsoleApp
{
  class Program
  {
    static void Main(string[] args)
    {
      Console.WriteLine("Hello MobiledgeX GRPC Library Sample App!");


      var mexGrpcLibApp = new MexGrpcLibApp();
      mexGrpcLibApp.RunSampleFlow();
    }
  }

  public class TokenException : Exception
  {
    public TokenException(string message)
        : base(message)
    {
    }

    public TokenException(string message, Exception innerException)
        : base(message, innerException)
    {
    }
  }

  class MexGrpcLibApp
  {
    Loc location;
    string sessionCookie;

    string dmeHost = "mexdemo.dme.mobiledgex.net"; // DME server hostname or ip.
    int dmePort = 50051; // DME port.

    MatchEngineApi.MatchEngineApiClient client;

    public void RunSampleFlow()
    {
      location = getLocation();
      string uri = dmeHost + ":" + dmePort;
      string devName = "MobiledgeX";
      string appName = "MobiledgeX SDK Demo";

      // Channel:
      ChannelCredentials channelCredentials = new SslCredentials();
      Channel channel = new Channel(uri, channelCredentials);

      client = new DistributedMatchEngine.MatchEngineApi.MatchEngineApiClient(channel);

      var registerClientRequest = CreateRegisterClientRequest(getCarrierName(), devName, appName, "1.0", "");
      var regReply = client.RegisterClient(registerClientRequest);

      Console.WriteLine("RegisterClient Reply Status: " + regReply.Status);
      Console.WriteLine("RegisterClient TokenServerURI: " + regReply.TokenServerUri);

      // Store sessionCookie, for later use in future requests.
      sessionCookie = regReply.SessionCookie;

      // Request the token from the TokenServer:
      string token = null;
      try
      {
        token = RetrieveToken(regReply.TokenServerUri);
        Console.WriteLine("VerifyLocation pre-query TokenServer token: " + token);
      }
      catch (System.Net.WebException we)
      {
        Debug.WriteLine(we.ToString());

      }
      if (token == null)
      {
        return;
      }


      // Call the remainder. Verify and Find cloudlet.

      // Async version can also be used. Blocking:
      var verifyResponse = VerifyLocation(token);
      Console.WriteLine("VerifyLocation Status: " + verifyResponse.GpsLocationStatus);
      Console.WriteLine("VerifyLocation Accuracy: " + verifyResponse.GpsLocationAccuracyKm);

      // Blocking GRPC call:
      var findCloudletResponse = FindCloudlet();
      Console.WriteLine("FindCloudlet Status: " + findCloudletResponse.Status);
      Console.WriteLine("FindCloudlet Response: " + findCloudletResponse);
    }


    RegisterClientRequest CreateRegisterClientRequest(string carrierName, string devName, string appName, string appVersion, string authToken)
    {
      var request = new RegisterClientRequest
      {
        Ver = 1,
        CarrierName = carrierName,
        DevName = devName,
        AppName = appName,
        AppVers = appVersion,
        AuthToken = authToken
      };
      return request;
    }

    VerifyLocationRequest CreateVerifyLocationRequest(string carrierName, Loc gpsLocation, string verifyLocationToken)
    {
      var request = new VerifyLocationRequest
      {
        Ver = 1,
        SessionCookie = sessionCookie,
        CarrierName = carrierName,
        GpsLocation = gpsLocation,
        VerifyLocToken = verifyLocationToken
      };
      return request;
    }

    FindCloudletRequest CreateFindCloudletRequest(string carrierName, Loc gpsLocation)
    {
      var request = new FindCloudletRequest
      {
        Ver = 1,
        SessionCookie = sessionCookie,
        CarrierName = carrierName,
        GpsLocation = gpsLocation
      };
      return request;
    }

    static String parseToken(String uri)
    {
      string[] uriandparams = uri.Split('?');
      if (uriandparams.Length < 1)
      {
        return null;
      }
      string parameterStr = uriandparams[1];
      if (parameterStr.Equals(""))
      {
        return null;
      }

      string[] parameters = parameterStr.Split('&');
      if (parameters.Length < 1)
      {
        return null;
      }

      foreach (string keyValueStr in parameters)
      {
        string[] keyValue = keyValueStr.Split('=');
        if (keyValue[0].Equals("dt-id"))
        {
          string value = null;
          int pos = keyValue[0].Length + 1;
          if (pos < keyValueStr.Length)
          {
            value = keyValueStr.Substring(pos, keyValueStr.Length - pos);
          }
          return value;
        }
      }

      return null;
    }

    string RetrieveToken(string tokenServerURI)
    {
      HttpWebRequest httpWebRequest = (HttpWebRequest)WebRequest.Create(tokenServerURI);
      httpWebRequest.AllowAutoRedirect = false;

      HttpWebResponse response = null;
      string token = null;
      string uriLocation = null;
      // 303 See Other is behavior is different between standard C#
      // and what's potentially in Unity.
      try
      {
        response = (HttpWebResponse)httpWebRequest.GetResponse();
        if (response != null)
        {
          if (response.StatusCode != HttpStatusCode.SeeOther)
          {
            throw new TokenException("Expected an HTTP 303 SeeOther.");
          }
          uriLocation = response.Headers["Location"];
        }
      }
      catch (System.Net.WebException we)
      {
        response = (HttpWebResponse)we.Response;
        if (response != null)
        {
          if (response.StatusCode != HttpStatusCode.SeeOther)
          {
            throw new TokenException("Expected an HTTP 303 SeeOther.", we);
          }
          uriLocation = response.Headers["Location"];
        }
      }

      if (uriLocation != null)
      {
        token = parseToken(uriLocation);
      }
      return token;
    }

    VerifyLocationReply VerifyLocation(string token)
    {
      var verifyLocationRequest = CreateVerifyLocationRequest(getCarrierName(), getLocation(), token);
      var verifyResult = client.VerifyLocation(verifyLocationRequest);
      return verifyResult;
    }

    FindCloudletReply FindCloudlet()
    {
      // Create a synchronous request for FindCloudlet using RegisterClient reply's Session Cookie (TokenServerURI is now invalid):
      var findCloudletRequest = CreateFindCloudletRequest(getCarrierName(), getLocation());
      var findCloudletReply = client.FindCloudlet(findCloudletRequest);

      return findCloudletReply;
    }

    // TODO: The app must retrieve form they platform this case sensitive value before each DME GRPC call.
    // The device is potentially mobile and may have data roaming.
    String getCarrierName() {
      return "TDG";
    }

    // TODO: The client must retrieve a real GPS location from the platform, even if it is just the last known location,
    // possibly asynchronously.
    Loc getLocation()
    {
      return new DistributedMatchEngine.Loc
      {
        Longitude = -122.149349,
        Latitude = 37.459609
      };
    }

  }
}
