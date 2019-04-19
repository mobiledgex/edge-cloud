using System;
using System.Diagnostics;
using System.IO;
using System.Net;
using System.Text;

using System.Net.Http;
using System.Runtime.Serialization.Json;
using System.Security.Cryptography.X509Certificates;

using System.Threading.Tasks;

namespace DistributedMatchEngine
{
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

  // Minimal logger without log levels:
  static class Log
  {
    // Stdout:
    public static void S(string msg)
    {
      Console.WriteLine(msg);
    }
    // Stderr:
    public static void E(string msg)
    {
      TextWriter errorWriter = Console.Error;
      errorWriter.WriteLine(msg);
    }

    // Stdout:
    [ConditionalAttribute("DEBUG")]
    public static void D(string msg)
    {
      Console.WriteLine(msg);
    }

  }

  public class MatchingEngine
  {
    public const UInt32 defaultDmeRestPort = 38001;
    public const string carrierNameDefault = "tdg";

    string baseDmeHost = "dme.mobiledgex.net";
    UInt32 dmePort { get; set; } = defaultDmeRestPort; // HTTP REST port

    // API Paths:
    private string registerAPI = "/v1/registerclient";
    private string verifylocationAPI = "/v1/verifylocation";
    private string findcloudletAPI = "/v1/findcloudlet";
    private string getlocationAPI = "/v1/getlocation";
    private string appinstlistAPI = "/v1/getappinstlist";
    private string dynamiclocgroupAPI = "/v1/dynamiclocgroup";
    private string getfqdnlistAPI = "/v1/getfqdnlist";

    UInt32 timeoutSec = 5000;
    string appName;
    string devName;
    string appVersionStr = "1.0";

    public string sessionCookie { get; set; }
    string tokenServerURI;
    string authToken { get; set; }

    public MatchingEngine()
    {
    }

    private HttpWebRequest createWebRequest(string uri)
    {
      HttpWebRequest httpWebRequest = (HttpWebRequest)WebRequest.Create(uri);
      // FIXME: This should not be here.
      // Server Cert: (self signed!)
      byte[] caCrtBytes = Encoding.ASCII.GetBytes(Credentials.caCrt);
      var x509CertServer = new X509Certificate2(caCrtBytes);
      httpWebRequest.ClientCertificates.Add(x509CertServer);

      // ClientCert:
      byte[] clientCredentials = Convert.FromBase64String(Credentials.clientCrtBase64); // Pkcs12 binary source.
      var x509ClientCertKeyPair = new X509Certificate2(clientCredentials, "foo"); /* 2nd param: FIXME: password for credentials */
      httpWebRequest.ClientCertificates.Add(x509ClientCertKeyPair);
      Log.D("Has Private Key?" + x509ClientCertKeyPair.HasPrivateKey);

      // FIXME: Real certs required.
      httpWebRequest.ServerCertificateValidationCallback = (sender, cert, chain, sslPolicyErrors) =>
      {
        Log.D("==== Sender: " + sender);
        Log.D("==== Cert: " + cert);
        Log.D("==== Chain: " + chain);
        Log.D("==== SSLPolicyErrors: " + sslPolicyErrors);
        return true;
      };

      return httpWebRequest;
    }

    public string getCarrierName()
    {
      return carrierNameDefault;
    }

    string generateDmeHostPath(string carrierName)
    {
      if (carrierName == null || carrierName == "")
      {
        return carrierNameDefault + "." + baseDmeHost;
      }
      return carrierName + "." + baseDmeHost;
    }

    public string GenerateDmeBaseUri(string carrierName, UInt32 port = defaultDmeRestPort)
    {
      return "https://" + generateDmeHostPath(carrierName) + ":" + port;
    }

    public string CreateUri(string host, UInt32 port)
    {
      if (host != null && host != "")
      {
        return "https://" + host + ":" + port;
      }
      return GenerateDmeBaseUri(null, port);
    }

    /*
     * This is temporary, and must be updated later.   
     */
    private bool SetCredentials(string caCert, string clientCert, string clientPrivKey)
    {
      return false;
    }

    private async Task<Stream> PostRequest(string uri, string jsonStr)
    {
      return await Task.Run(() =>
      {
        var webRequest = createWebRequest(uri);
        // Choose network TBD (async)

        webRequest.Method = "POST";
        webRequest.ContentType = "application/json";
        webRequest.ContentLength = jsonStr.Length;

        // Write the data:
        using (var outstream = webRequest.GetRequestStream())
        {
          using (StreamWriter streamWriter = new StreamWriter(outstream))
          {
            streamWriter.Write(jsonStr);
          }
        }

        Stream replyStream = webRequest.GetResponse().GetResponseStream();
        return replyStream;
      });
    }

    private static String parseToken(String uri)
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
        int pos = keyValue[0].Length + 1; // step over '='
        if (pos >= keyValueStr.Length)
        {
          return null;
        }

        string value = keyValueStr.Substring(pos, keyValueStr.Length - pos);
        if (keyValue[0].Equals("dt-id"))
        {
          return value;
        }
      }

      return null;
    }

    private string RetrieveToken(string aTokenServerURI)
    {
      HttpWebRequest httpWebRequest = (HttpWebRequest)WebRequest.Create(aTokenServerURI);
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
        Log.D("uriLocation: " + uriLocation);
        token = parseToken(uriLocation);
      }

      if (token == null)
      {
        throw new InvalidTokenServerTokenException("Token not found or parsable in the URI string: " + uriLocation);
      }

      return token;
    }

    public RegisterClientRequest CreateRegisterClientRequest(string carrierName, string developerName, string appName, string appVersion, string authToken)
    {
      return new RegisterClientRequest {
        Ver = 1,
        CarrierName = carrierName,
        AppName = appName,
        AuthToken = authToken,
        AppVers = appVersionStr,
        DevName = developerName };
    }

    public async Task<RegisterClientReply> RegisterClient(string host, uint port, RegisterClientRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(RegisterClientRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + registerAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(RegisterClientReply));
      RegisterClientReply reply = (RegisterClientReply)deserializer.ReadObject(responseStream);
      this.sessionCookie = reply.SessionCookie;
      this.tokenServerURI = reply.TokenServerURI;
      return reply;
    }

    public FindCloudletRequest CreateFindCloudletRequest(string carrierName, string devName, string appName, string appVers, Loc loc)
    {
      if (sessionCookie == null)
      {
        // Exceptions.
        return null;
      }
      return new FindCloudletRequest
      {
        SessionCookie = this.sessionCookie,
        CarrierName = carrierName,
        DevName = devName,
        AppName = appName,
        AppVers = appVers,
        GpsLocation = loc
      };
    }

    public async Task<FindCloudletReply> FindCloudlet(string host, uint port, FindCloudletRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(FindCloudletRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + findcloudletAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(FindCloudletReply));      
      FindCloudletReply reply = (FindCloudletReply)deserializer.ReadObject(responseStream);
      return reply;
    }


    public VerifyLocationRequest CreateVerifyLocationRequest(string carrierName, Loc loc)
    {
      if (sessionCookie == null)
      {
        return null;
      }
      return new VerifyLocationRequest {
        Ver = 1,
        CarrierName = carrierName,
        GpsLocation = loc,
        SessionCookie = this.sessionCookie,
        VerifyLocToken = null
      };
    }

    public async Task<VerifyLocationReply> VerifyLocation(string host, uint port, VerifyLocationRequest request)
    {
      string token = RetrieveToken(tokenServerURI);
      request.VerifyLocToken = token;

      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(VerifyLocationRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + verifylocationAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(VerifyLocationReply));
      VerifyLocationReply reply = (VerifyLocationReply)deserializer.ReadObject(responseStream);
      return reply;
    }

    public GetLocationRequest CreateGetLocationRequest(string carrierName)
    {
      if (sessionCookie == null)
      {
        return null;
      }
      return new GetLocationRequest
      {
        Ver = 1,
        CarrierName = carrierName,
        SessionCookie = this.sessionCookie
      };
    }

    /*
     * Retrieves the carrier based network based geolocation of the network device.
     */
    public async Task<GetLocationReply> GetLocation(string host, uint port, GetLocationRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(GetLocationRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + getlocationAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(GetLocationReply));
      GetLocationReply reply = (GetLocationReply)deserializer.ReadObject(responseStream);
      return reply;
    }


    public AppInstListRequest CreateAppInstListRequest(string carrierName, Loc loc)
    {
      if (sessionCookie == null)
      {
        return null;
      }
      if (loc == null)
      {
        return null;
      }

      return new AppInstListRequest
      {
        Ver = 1,
        CarrierName = carrierName,
        SessionCookie = this.sessionCookie,
        GpsLocation = loc
      };
    }

    public async Task<AppInstListReply> AppInstList(string host, uint port, AppInstListRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(AppInstListRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + appinstlistAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(AppInstListReply));
      AppInstListReply reply = (AppInstListReply)deserializer.ReadObject(responseStream);
      return reply;
    }

    public FqdnListRequest CreateFqdnListRequest()
    {
      if (sessionCookie == null)
      {
        return null;
      }

      return new FqdnListRequest
      {
        Ver = 1,
        SessionCookie = this.sessionCookie
      };
    }

    public async Task<FqdnListReply> FqdnList(string host, uint port, FqdnListRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(FqdnListRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + getfqdnlistAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(FqdnListReply));
      FqdnListReply reply = (FqdnListReply)deserializer.ReadObject(responseStream);
      return reply;
    }

    public DynamicLocGroupRequest CreateDynamicLocGroupRequest(UInt64 lgId, DlgCommType dlgCommType, string userData)
    {
      if (sessionCookie == null)
      {
        return null;
      }

      return new DynamicLocGroupRequest
      {
        Ver = 1,
        SessionCookie = this.sessionCookie,
        LgId = lgId,
        CommType = dlgCommType.ToString(), // JSON REST request needs a string, not integer-like type.
        UserData = userData
      };
    }

    public async Task<DynamicLocGroupReply> DynamicLocGroup(string host, uint port, DynamicLocGroupRequest request)
    {
      DataContractJsonSerializer serializer = new DataContractJsonSerializer(typeof(DynamicLocGroupRequest));
      MemoryStream ms = new MemoryStream();
      serializer.WriteObject(ms, request);
      string jsonStr = Util.StreamToString(ms);

      Stream responseStream = await PostRequest(CreateUri(host, port) + dynamiclocgroupAPI, jsonStr);
      if (responseStream == null || !responseStream.CanRead)
      {
        return null;
      }

      DataContractJsonSerializer deserializer = new DataContractJsonSerializer(typeof(DynamicLocGroupReply));
      DynamicLocGroupReply reply = (DynamicLocGroupReply)deserializer.ReadObject(responseStream);
      return reply;
    }

  };

  static class Credentials
  {
    // Root CA:
    public static string caCrt = @"-----BEGIN CERTIFICATE-----
MIIE4jCCAsqgAwIBAgIBATANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZtZXgt
Y2EwHhcNMTgwODIwMDMwNzQwWhcNMjAwMjIwMDMwNzQwWjARMQ8wDQYDVQQDEwZt
ZXgtY2EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCk45wuENmZk/ok
u41JjBwC3PBRWA8SRIGjVbHHIGA6uORNY/GaKU1mgBengzvOqT9DrnwsHjQGoLEl
f3J3d/KF3rhm60ZVtHGi3FUuZc9N/8E3ABBivd31gGcv30b25UyxcE9TNiqJJ7z2
t1RmLT5+KSP7Mg9l+H1lhBRukdmAIym+pQsiFaKTxiZ9VbCfget+QP4mn1G1VdLD
j9LV7UIS8OO3EwslU4yYT0ycaHoUZEKWUYu9+D+nL/L9X3lruk2vNg0ieiFBAeoM
hAmQy8sQ+nsLPQlzWr7nyi2AMCwL8DQyKFVkGETaCmmTrDhoy+TWE6wkX5nsw90t
EIaaE4/BpEc4irXT8lUdLXEovYWaFcGxTDMJkFKjtPFACO8ck8r0GQ1C24FWqU4B
7kQoCj9BRtTGfRAozBIOroZ2/k2lig+FGIVnSUPIvc0F2axFgzlg8k2RVrPrdkQ2
VYWG7FLfZ9Wo8SkkenmnejguBEKOXs0Q4kEjh/qjDx5MZgIvjRCiamUt2MnBm7ky
j1EA1lZ7A8f2FmbCLthvcu5knpI1w/yvADXC2i+dBa/r8pIIlZv/9Mb7IlkBBF95
p0ES3rXAMTVbTrYKnwPSiZoES5zuLHOKCrNLg2zO2nqYE69a6FEPjkA+aoOs/KZG
JauZzTvtP25hViMnuee9QckcwS/mzwIDAQABo0UwQzAOBgNVHQ8BAf8EBAMCAQYw
EgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU0JOPAC7G5mw7G7W0e6GtHn7r
S3gwDQYJKoZIhvcNAQELBQADggIBAIDFI/SbMNIy1nskp4TNKv6YwMMgWUO7tfXs
obLrwGfneOR8lA9GJlpab47aohWcTxua6iwzUNqowq0x5wwmWwbSLeyiMY4TJj2E
C6Lla7uuC6WeY3RIS1XjjvOIvW6Mq2n3JElIwtGnm5qmr6CzfqiZenxY+UU1nBbI
eRw/V36ikJAQz5kj7wskUVwhAPEPnHr1hYmyu8t2Ue7zERwHzRIs0nwERQaV32aI
f//bV/kqhzueDiwqXqwFSHGreEefhUGUYEtiC8etLUZHe5ts3627pTjr0FZ2F+pt
OkQB9A+yuTPeJQJTMCLuyHxsiDQkT0On/Mky6ffdWSAwWXbcNVfO1If8Wi6yXvRz
Xe9dvyiIVwvG4VAtuwGEmTXd/fh4J/OpqxjLcXe/3k1ibX+y6zClV4REp4ygm6Kr
minVSTY4o4f1H39tIth9LZqpZKzOC4QJu0CWI4rLb6sCUlo8nWOmAktyua3xSAjs
k59JzlMIfdZ8z0SZq0Hy5Bf5XhZY0WcvWLc89RGslDibMtg6qweO43F23lV8w+Xn
mbSi8TUL2D7kA5StYElnJ4G2o4Bmymu8XxcZhDfeH0LJ8lqP7TyRnkL2jmNYm6be
3u/yuyFCrwRluMzwEzAY+3FPuhfHWCmlSZhx1gsXQIXtmKT0l2xU9dlF8fPBYC5e
LggXHNeu
-----END CERTIFICATE-----";

    // Client pkcs12 binary bytes with cert and private key in Base64 format:
    public static string clientCrtBase64 = "MIIK2gIBAzCCCqAGCSqGSIb3DQEHAaCCCpEEggqNMIIKiTCCBQ8GCSqGSIb3DQEHBqCCBQAwggT8AgEAMIIE9QYJKoZIhvcNAQcBMBwGCiqGSIb3DQEMAQYwDgQIzDLW8crne/QCAggAgIIEyEFVVloxzCbtO1zBZr8s9tejKtM+ffd1M/eyJff5lAYuSUFivhH3FKFfYaZV+klOMETNt876y+Q4vVHfU/MC5aRA4QLf04EHj+TPJ4WiTRynemKOw1eT3QsoCIGsQy8O77qz8UvtO1ezBThfkcw84fIqDa2ngaYt1MLblWY8JOUmo+v8TNhj7OGF3Iz9HbN3Jpxo2urbaKtBfnAjKAmycyjZqp780xWvE4ZTExHf+wi+NARhhLSLWt2oc6zxkF63R3LtEzM0YLdsUpKc9SMU7sEl0P2Po4YUiNmf2GSkDkuq0T2V6vtW1ESFj++daeDk6GLsh8puAYp2lwOKpTayj7wHent272lW2nRoeGcvW2cmEpBN4umvACr0nJ7WWRD9cktbf8nsmursK8ezor8I+TWlF4GB/J37LWEbQfbEf6E0QpL3I1UXQsrfrqA0h1jIEax9zkyR9ryHk4DkTFyPlm9LlSkxiqotSmcatX6JWANT99Vh1m8iUTwSUkJSin+5Lylh5n99DeQ2x8+ltpd+Na2Zj5Cj9Nq/0JuLi7dWkYNOofE6LwbKSIY++mSXD5PimomrANLlGnf8bcpE2tJUch+Dhigrpf/N8KWi0YXR8euy8iPICEN8SZHO7WxdUwoYxAAZk0mlmlMCynTxSxKxiuQgc2YzWkF3lExDVDFiyZbVGzhBSIUGBjiuwVOC1lw4yCHcB1P4pG9B5LCoUgRBH99+VWsWN3aOqlSIRkzJhjri3491ORsR8lFgRy2RL0X5UJtH21DkaJuXWrQo0Glv5CPOREVmZyktUZQofg330m5hfL17geaU6e5jZ8HYbV/c2GzBChFT9S320URh4FRqYEOT0rBmasid3NV5TrxnmGD7gMhcWcdl8CeCspTIkrxjOyZQJFqKVVNcquqvNqR9ACOO2nC1HBEfok4BQQvFVLpb328K8Ecc3V6GTQSoBoAIygJZ2ioIWkKVGV+bVPyDaGdGkJUlceH50iVzunPwtyFovdPkgcWIt8V87wXxg3t1aeBo1deEppAMQq1/DmEdXvIF5C/KFr6qHf38pUH2KFTl0jjgicWCBNdXNJO1iijBgwJ7hKJ9fboc6Qtk057ISvQZ7VL9gi1AFMpM8mCNmj3w2pkxTxDarcvHvOIEhW3Ap+XgSe1S5TAZTAUOkxA/flnBNzNwXXKJ4B1srfb+AMLVvaEPothAllYUGZq2PoK+1/H9yvBxl6e79PjutJpCMuIXsR0NtXafNAVjcsPT2Bekbz4famzXBzFR/FMGqFYknFCAH8XA12YmYrIqipxGD4AiClI5vTvT3GIa8hubSY7o8lHBrPFvw4xUjy5z6homQ1HZJkC/nKp4ZZ8Y5KMkVF5IzHs1O4Aa5B+r3VGp103dsIHEHpPzJSNIezxjTKejlR9MSI9TAmNQ8mKmp83hUDRvEoku532lkoS0DQPJdEvS7HlWKiAuDvZ/3lWf1YZVXtzBJo/xBsUU/C58gQsgWmMDSAUuykaUj/86bkk/j5ePT1nTm5G6gGo523TAAfvIbrbnDl4+d9tGHb/Fg813i8Rm/Zyar/1I69uZvEDmzq8LFiptavx6x3J+SdOAfnX9PcnOlQQPQMlzlXstjduGibc1XnvSVzf0JTCCBXIGCSqGSIb3DQEHAaCCBWMEggVfMIIFWzCCBVcGCyqGSIb3DQEMCgECoIIE7jCCBOowHAYKKoZIhvcNAQwBAzAOBAiW4HYGKKy8zQICCAAEggTIm4lGAMwHsPOCxw91lC1ZYK+Y0Nvj6jKmme+JrVGgItQ3TTA13/h0QddWQ5TLq90lMD6ADz7yNgKooDOSTxq9W8NBjJ6RlTFM+9v0phTaRWNwAuUV6vLp3xn/YC/Ig9020WZ6O8TOTXbfoIgmeD0QOifqGYfMxbUr3nvKZ8eYGU4AJgFzx3u3CMx/VMohKq055GvU0OAqWk2KCGCVO6tdEg8w9AvqRlKMlzD3JgNc0+NZwLEdYTlcgEWjYFXBP9xxdZ/SfL6uNiaonWgteePocE2/8uJvmPbFyiQ9AAXMRjCExPlwFHcttuB2MWsu55RY5ETHT7lrScRnyRKiOpWEkdhLuVdX8Zy6J32xTvG3FUVkj7SVDRR9f622A92bK1XuyaKas+x7zaJPiUxeHv3uk0XoXwu3ibchT2EcEkOLZCyvuq2j6NoxV8fso/YmAaCv56eB1XxkoHfr4Pcyhc5ops8ZfayM09V1geskz2aZySYzy4goJnOgHJNRyFFan1Izb2fAp4yVUPkXmudRrBLV2/26pOkGK6YXcyYWFEhejeF31+j4lQcRVoJKlU/AXhhDgqtLgJ0V+sIFE08jtcmrd0dETge7n2dD4Udj745vwGhHKMin8ccL9UIt6efxyZLTBIPRzpJWSqucnMkaobqgt9eFozmJjonEC9Y47Hdh0gy/asd8RzhUf5sBs8l3XIDFxf8E/bz1f5QffzOBojz4qgISxNQCYIM8JG5w58B05wwTzO6A1Bj/h9ByswLpgPguJt2FUcCY3ItmNcMLHZM6u6j1zWR/8koDWBjV/GEx8JylkUEoMHlq3JCKJRq2Z4LIAzd/DXcw3wdGfK6CAtspzhgGFbb4fuKEInlAs16QPGRFq9k4GRHuIp0Hi3HVSBmtUCbspzAITs4UK7mWMVM3LzcJxrzHMzjHzVS3TOfXEg42ywFtm+81z/VcDtuxqaFvb2g69VqBOsbVK4hA/PgtDFLm3AXNL3cqIc9K96rI/OuHsdcnCTk8VGgsj5lJc2RNFp4K3RAwfSM0tzd5g2nld+q5nSivWBRjB3Qxt7zTraMW2FEAjRWWj+sRdEHQbDJvXWS3+ez1m5j5/sd3KhpEOvOO5Vd2JlpEsnkvEMnm5Dob6zzWr3Z722OEN9TscCObe03tMtE4xifTEEUZERdqa9ELaZjF+DyOkNK72BKvlfJhs/+uKQF/Cb6Tf2fYzjVYPZTMWaFeW04NeHPGg/TJCg+jjhTtFFpIkevxI2JFN6DLmwe27kM+//DMf8n6l/FB+IZ0MZmmAbF8VCSNg2hTVbXzshSUEYpIaW8N6JoGjW7lQx9Hue3b7Pf0VBzHNsz7M/OPHVYkczZNh/LWV3IeDoTBlZbm3L227/V6splX0DqjmoX6jNgjFhi5a9OD2pmZZq/lWpSKRpDJm8yA/v/21NdrKB9MtmCQs8taQ2dfeAVElaZd2vi1Jb2RR8eIYgkABqy52CTxO3VDfUkRq62MB3uSxX/9CBh+JThAv7T/EPr9K/ZjvwgbG/vPetnQy08Y/3els/A604tpljGtcadvXC7T1atVvjNMp9on0WwIvgT87WDmmX3Wnn0GtfVh0vAKSnMZasYQFxc09tp9jAvY30ACXAbxTzqVMVYwIwYJKoZIhvcNAQkVMRYEFORaIqgsquac74x48KQZKR+qgmEOMC8GCSqGSIb3DQEJFDEiHiAA4gCAAJwAbQBlAHgALQBjAGwAaQBlAG4AdADiAIAAnTAxMCEwCQYFKw4DAhoFAAQUENIw9Bh6+cvCUxkkoB51xl9FV3kECPjGRQnsv7rwAgIIAA==";
  };
}
