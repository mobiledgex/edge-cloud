using UnityEngine;

using Grpc.Core;
using System;
using System.Net;
using System.Text;

using DistributedMatchEngine;

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

public class MexGrpcSample : MonoBehaviour
{
    string sessionCookie;
    string token;
    string authToken = ""; // MEX Developer supplied and updated authToken.

    string dmeHost = "tdg2.dme.mobiledgex.net"; // DME server hostname or ip.
    int dmePort = 50051; // DME port.

    string developerName = "EmptyMatchEngineApp";
    string applicationName = "EmptyMatchEngineApp";
    string appVer = "1.0";

    Match_Engine_Api.Match_Engine_ApiClient client;

    StatusContainer statusContainer;

    // Use this for initialization
    void Start()
    {
        Environment.SetEnvironmentVariable("GRPC_TRACE", "api");
        Environment.SetEnvironmentVariable("GRPC_VERBOSITY", "debug");
        Grpc.Core.GrpcEnvironment.SetLogger(new Grpc.Core.Logging.ConsoleLogger());
        statusContainer = GameObject.Find("/UICanvas/SampleOutput").GetComponent<StatusContainer>();
        RunSampleFlow();
    }

    // Update is called once per frame
    void Update()
    {

    }


    public void RunSampleFlow()
    {
        string uri = dmeHost + ":" + dmePort;

        // Channel:
        // TODO: Load from file or iostream, securely generate keys, etc.
        var clientKeyPair = new KeyCertificatePair(Credentials.clientCrt, Credentials.clientKey);
        var sslCredentials = new SslCredentials(Credentials.caCrt, clientKeyPair);
        Channel channel = new Channel(uri, sslCredentials);

        client = new Match_Engine_Api.Match_Engine_ApiClient(channel);

        var registerClientRequest = CreateRegisterClientRequest(developerName, applicationName, appVer, getCarrierName(), "");
        var regReply = client.RegisterClient(registerClientRequest);
        this.authToken = registerClientRequest.AuthToken; // Developer supplied.

        Console.WriteLine("RegisterClient Reply: " + regReply);

        Console.WriteLine("RegisterClient TokenServerURI: " + regReply.TokenServerURI);
        statusContainer.Post("RegisterClient TokenServerURI: " + regReply);
        this.authToken = registerClientRequest.AuthToken;


        // Store sessionCookie, for later use in future requests.
        sessionCookie = regReply.SessionCookie;

        // Request the token from the TokenServer:
        this.token = RetrieveToken(regReply.TokenServerURI);

        statusContainer.Post("VerifyLocation pre-query sessionCookie: " + sessionCookie);
        statusContainer.Post("VerifyLocation pre-query TokenServer token: " + token);

        // Call the remainder. Verify and Find cloudlet.

        // Async versions also exist:
        var verifyResponse = VerifyLocation();
        Console.WriteLine("VerifyLocation Status: " + verifyResponse.GpsLocationStatus);
        Console.WriteLine("VerifyLocation Accuracy: " + verifyResponse.GPSLocationAccuracyKM);
        statusContainer.Post("VerifyLocation Status: " + verifyResponse.GpsLocationStatus);
        statusContainer.Post("VerifyLocation Accuracy: " + verifyResponse.GPSLocationAccuracyKM);
        var vrSb = new StringBuilder();
        vrSb.Append("VerifyLocation Status:")
            .Append(", Status: " + verifyResponse.ToString());
        statusContainer.Post(vrSb.ToString());

        // Straight blocking GRPC call:
        var findCloudletResponse = FindCloudlet();
        Console.WriteLine("FindCloudlet Status: " + findCloudletResponse.Status);
        Console.WriteLine("FindCloudlet FQDN Location: " + findCloudletResponse.FQDN);

        var strBuilder = new StringBuilder();
        strBuilder.Append("FindCloudlet: ")
                  .Append(", Version: " + findCloudletResponse.Ver)
                  .Append(", Location Found Status: " + findCloudletResponse.Status)
                  .Append(", Location of cloudlet. Latitude: " + findCloudletResponse.CloudletLocation.Latitude)
                  .Append(", Longitude: " + findCloudletResponse.CloudletLocation.Longitude)
                  .Append(", Cloudlet FQDN: " + findCloudletResponse.FQDN + "\n");

        Google.Protobuf.Collections.RepeatedField<AppPort> ports = findCloudletResponse.Ports;
        foreach (AppPort appPort in findCloudletResponse.Ports)
        {
            strBuilder.Append(", AppPort: Protocol: " + appPort.Proto)
                      .Append(", AppPort: Internal Port: " + appPort.InternalPort)
                      .Append(", AppPort: Public Port: " + appPort.PublicPort)
                      .Append(", AppPort: Public Path: " + appPort.PublicPath);
        }
        statusContainer.Post(strBuilder.ToString());

    }

    RegisterClientRequest CreateRegisterClientRequest(string devName, string appName, string appVers, string carrierName, string anAuthToken)
    {
        var request = new RegisterClientRequest
        {
            Ver = 1,
            DevName = devName,
            AppName = appName,
            AppVers = appVers,
            CarrierName = carrierName,
            AuthToken = anAuthToken == "" ? authToken : anAuthToken
        };
        return request;
    }

    VerifyLocationRequest CreateVerifyLocationRequest(string carrierName, Loc gpsLocation)
    {
        if (this.token == null) {
            Console.WriteLine("Missing TOKEN!!!!!!");
        }
        var request = new VerifyLocationRequest
        {
            Ver = 1,
            SessionCookie = sessionCookie,
            CarrierName = carrierName,
            GpsLocation = gpsLocation,
            VerifyLocToken = this.token
        };
        return request;
    }

    FindCloudletRequest CreateFindCloudletRequest(string carrierName, Loc gpsLocation)
    {
        var request = new FindCloudletRequest
        {
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
                return keyValue[1];
            }
        }

        return null;
    }

    string RetrieveToken(string tokenServerURI)
    {
        HttpWebRequest httpWebRequest = (HttpWebRequest)WebRequest.Create(tokenServerURI);
        httpWebRequest.AllowAutoRedirect = false;

        HttpWebResponse response = null;
        string atoken = null;
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
            atoken = parseToken(uriLocation);
        }
        return atoken;
    }


    VerifyLocationReply VerifyLocation()
    {
        // Use verifyLocation, Async call:
        var verifyLocationRequest = CreateVerifyLocationRequest(getCarrierName(), getLocation());
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

    // This could change on every tower change.
    string getCarrierName() {
        return "TDG";
    }

    Loc getLocation()
    {
        // TODO
        return new DistributedMatchEngine.Loc
        {
            Longitude = -122.149349,
            Latitude = 37.459609,
            Altitude = 0,
            Course = 0,
            HorizontalAccuracy = 0,
            VerticalAccuracy = 0,
            Speed = 5,
            Timestamp = new DistributedMatchEngine.Timestamp {
                Nanos = 10000,
                Seconds = 1000
            }
        };
    }
}

// Test Certs ONLY.
class Credentials
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

    // Client Crt:
    public static string clientCrt = @"-----BEGIN CERTIFICATE-----
MIIEOjCCAiKgAwIBAgIQMCuDiDXhpNOKRvn69uSbCjANBgkqhkiG9w0BAQsFADAR
MQ8wDQYDVQQDEwZtZXgtY2EwHhcNMTgwODIwMDMxMDI0WhcNMjAwMjIwMDMwNzM5
WjAVMRMwEQYDVQQDEwptZXgtY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAzLaPAPOAUzV49VXgiYsDZTQ/zyCtsr0w3Ge/tvIck2Mm2FtAZ88r
oRV2UPrviEZPBL+o/JPiShfgqj1cLU1GyRr4uyezYl9AIig9/2xjYnkcXg6e3QG7
5lOaX2zH9mrYAm/N2hNzJwe9ZWBibXMDwFN4cptyygZuQ86SnK4j+6h1SVmN+F1j
ma18RzSU2rOjHK0InFgILSOlcjYYD/ds3HiL+vY6CVSNfuul5IMvQ/R8B8GwZpGG
34vl4AsDYM6FK+qW/ncFEt9rAOfu9AOeDWciKxPjzB4bX5G0dWAk5IU3VoGafEXe
Yt1xyeK25KHkw4kSh57BRkJKSuyvPWUqewIDAQABo4GJMIGGMA4GA1UdDwEB/wQE
AwIDuDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFGoT
cCzivXGPbeBD14sTusqjamkbMB8GA1UdIwQYMBaAFNCTjwAuxuZsOxu1tHuhrR5+
60t4MBUGA1UdEQQOMAyCCm1leC1jbGllbnQwDQYJKoZIhvcNAQELBQADggIBABSb
ZdZE9zh7vgT3cNr0gLVWk4We43VRlSa5JgdzWixi63qYTVIjHUE0BZo9yBMm3cx8
w327wAMVdlImUl/0Sv1NLi0IyC7EtxHfNhmUsgDa9oXuZwXM/RNmite1emS0ZYLE
fh7lNYUwnU/DHRPlZbIu08/7jYfyqxYKJSkP9HUBGOreGmMg+xARifr1Wb9uQDnx
12zK2uSG0Np8SbZV6FykNp7HeXG0jzOW+1Kg/NSxzYtScbSj3PfHIIHCiPhLd630
jVFsRUSfKMsM2+mDotVMvFlGyKoQUkQpS9RT/WsbKsSgLuW/WP1Zskh692DchMxI
YIASjJnwgidLwVen5uwj6h9vX9L1jb8CjV8w+SPjTqo6Hw8veQnqZV1nI1PgE17P
wjBlAmrCzfhbpNoMrktglWhpi/NJtCfQWwNkkGpd547NA1WXTRRqbxGc9Mg4GALb
6G9Hzp7nUpC3tC78jwZtsiw6oyPreYQc+O1XVD7qrvfcKabDStOIYcyenLmv82CL
ciCxLZGwVZ+5ABTuOsch7Sg3ZFnCJHvJqn00ZH7YBvTeoCfeelEiFCn7ehEbX8kn
y9MuVRwaDfGMKBYZ8Urr+RDxoDQwObvYQNIW9wVLoCGRuj3qt0BtTLkOuHvNCox1
GJJmmOKzRetckg3+ZSu7ET2YDuB3TDxPvDyc4Tq8
-----END CERTIFICATE-----";

    // Client Private Key:
    public static string clientKey = @"-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzLaPAPOAUzV49VXgiYsDZTQ/zyCtsr0w3Ge/tvIck2Mm2FtA
Z88roRV2UPrviEZPBL+o/JPiShfgqj1cLU1GyRr4uyezYl9AIig9/2xjYnkcXg6e
3QG75lOaX2zH9mrYAm/N2hNzJwe9ZWBibXMDwFN4cptyygZuQ86SnK4j+6h1SVmN
+F1jma18RzSU2rOjHK0InFgILSOlcjYYD/ds3HiL+vY6CVSNfuul5IMvQ/R8B8Gw
ZpGG34vl4AsDYM6FK+qW/ncFEt9rAOfu9AOeDWciKxPjzB4bX5G0dWAk5IU3VoGa
fEXeYt1xyeK25KHkw4kSh57BRkJKSuyvPWUqewIDAQABAoIBACePjBk59W2fIs3+
l5LdC33uV/p2LTsidqPRZOo85arR+XrMP6kQDzVlCWVi6RFjzPd09npBNfTtolwj
2YFjsq9AiBra9D6pe6JeNoT69EXec831c1vwbth3BZk1U3tacH4gDx76rUE4rLA/
rSXLmUj8mIVFZyyFi5+M9yZSPN/v+J7VfE2b3dRVw+oBa4GEGpR9IcEGyGgjzz/2
qc4e+iphCW10ZfLl6ZuVIErJgoGEDlilZhNMLlGIu/4/0Pd2DH+pYE4JaVKHIltW
WChDIWh1Ad+fZvVB/+M27GOaavHKm64z5/46Oj/n6lhwL/GDi1CQeMHBSFn4xsf0
rloEvNkCgYEA3rChHJlq7xxHnAyCjR+VVWjS409DQHFM2KntbX7mv2lMS8t4jpf9
sV/tyX5CWrlbk4Ntyxzz8Z8hi6VYhRsLVDbWL5g3fHMK76mj1JtbDLUw/w28twEm
d1wit0i+5Un6i3i7Wig8g+y31xAqobI4/5mhehEhjjSQP7H8JQ7M2dUCgYEA61WM
x5QGAzYIi+y59rFITfk6x7DNW7dOk0z3B6caF1HBoUzIh/uZrKaGxHLUcPNUprBG
vhSCpYdsXPVZgqxibnjPwrbYCPBZoN4DXsvOosbUcK4aDwvAX/i0RH77eBM5M+J4
5DbNxoS8y0ALjjgj85LQv4fFQeFXy2VtVGvHSw8CgYAyrTtcyMT++Q6KwoYLG37e
WuZy+Byz05TLUZBIdLKKKKpGLV2YBZqj/NKeIe9zue7PGP+pU0NoXvBBWTVVxRvE
5F3Fovwtg/ifJZm0zk3gDHPD9xpVAxv/2aXE0/ctMrKjfqwUDkgHNZ14gaNR/L7f
29RVdQSP2gJhnF1nCYEwqQKBgQDBd+Vytfhzb1p7XjRL4Ncmczylqm5Jdlt8sYts
mS3T+fyLlMpPMMLXs1eb7SNFcGYpW0XtQoNdfgXSLkpWKU4Kr/ttgk/8mUu1+o8e
wcKxA3Dm6dq2f9y5iYb5wMMPpg4i346vX3awO7PSDGbzlqfHuO0waHf8fztkFZBa
FPkUdQKBgGhCavU72BxbSIFBsLuw07+DQ7aU00JFkqEFYrK2Y0KbtRzi5s0f8sTQ
m67Ck9nZhhvdpunWWeynM8rjJy4MPUJWZeJmo8OAxOAdmdXs8cmykqfNb+0uC35b
i4CrX+qG6rq9Y4/kVJ4jWdUbpAN7gp+vCMBUGZ0HuYtxlkRH4y6G
-----END RSA PRIVATE KEY-----";
}