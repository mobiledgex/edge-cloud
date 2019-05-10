#include <curl/curl.h>

#include <sstream>
#include <iostream>

#include <nlohmann/json.hpp>

// Test only credentials.
#include "test_credentials.hpp"

using namespace std;
using namespace std::chrono;
using namespace nlohmann;

class MexRestClient {
  public:
    string carrierNameDefault = "TDG";
    string baseDmeHost = "dme.mobiledgex.net";
    unsigned int dmePort = 38001; // HTTP REST port

    // API Paths:
    string registerAPI = "/v1/registerclient";
    string verifylocationAPI = "/v1/verifylocation";
    string findcloudletAPI = "/v1/findcloudlet";
    string getlocatiyonAPI = "/v1/getlocation";
    string appinstlistAPI = "/v1/getappinstlist";
    string dynamiclocgroupAPI = "/v1/dynamiclocgroup";

    unsigned long timeoutSec = 5000;
    const string devName = "MobiledgeX"; // Your developer name
    const string appName = "MobiledgeX SDK Demo"; // Your application name
    const string appVersionStr = "1.0";

    // SSL files:
    unique_ptr<test_credentials> test_creds;
    string caCrtFile = "../../../tls/out/mex-ca.crt";
    string clientCrtFile = "../../../tls/out/mex-client.crt";
    string clientKeyFile = "../../../tls/out/mex-client.key";


    MexRestClient() {
        this->test_creds = unique_ptr<test_credentials>(
            new test_credentials(caCrtFile, clientCrtFile, clientKeyFile));
    }

    // Retrieve the carrier name of the cellular network interface.
    string getCarrierName() {
        return carrierNameDefault;
    }

    string generateDmeHostPath(string carrierName) {
        if (carrierName == "") {
            return carrierNameDefault + "." + baseDmeHost;
        }
        return carrierName + "." + baseDmeHost;
    }

    string generateBaseUri(const string &carrierName, unsigned int port) {
        stringstream ss;
        ss << "https://" << generateDmeHostPath(carrierName) << ":" << dmePort;
        return ss.str();
    }

    json currentGoogleTimestamp() {
        // REST Timestamp format based on Google's google.protobuf.Timestamp (RFC3339 string format)
        auto ts_micro = std::chrono::system_clock::now().time_since_epoch();
        auto ts_sec = duration_cast<std::chrono::seconds>(ts_micro);
        auto ts_micro_remainder = ts_micro % std::chrono::microseconds(1000000);
        auto ts_nano_remainder = duration_cast<std::chrono::nanoseconds>(ts_micro_remainder);

        json googleTimestamp;
        googleTimestamp["seconds"] = ts_sec.count();
        googleTimestamp["nanos"] = ts_nano_remainder.count();

        return googleTimestamp;
    }

    // A C++ GPS location provider/binding is needed here.
    json retrieveLocation() {
        json location;
        location["latitude"] = 37.459609;
        location["longitude"] = -122.149349;
        location["horizontal_accuracy"] = 5;
        location["vertical_accuracy"] = 20;
        location["altitude"] = 100;
        location["course"] = 0;
        location["speed"] = 2;
        location["timestamp"] = currentGoogleTimestamp();

        return location;
    }

    json createRegisterClientRequest() {
        json regClientRequest;

        regClientRequest["ver"] = 1;
        regClientRequest["AppName"] = appName;
        regClientRequest["DevName"] = devName;
        regClientRequest["AppVers"] = appVersionStr;
        regClientRequest["CarrierName"] = getCarrierName();
        regClientRequest["AuthToken"] = ""; // Developer supplied user auth token.

        return regClientRequest;
    }

    // Carrier name can change depending on cell tower.
    json createVerifyLocationRequest(const string &carrierName, const json gpslocation, const string verifyloctoken) {
        json verifyLocationRequest;

        verifyLocationRequest["ver"] = 1;
        verifyLocationRequest["SessionCookie"] = sessioncookie;
        verifyLocationRequest["CarrierName"] = carrierName;
        verifyLocationRequest["GpsLocation"] = gpslocation;
        verifyLocationRequest["VerifyLocToken"] = verifyloctoken;

        return verifyLocationRequest;
    }

    // Carrier name can change depending on cell tower.
    json createFindCloudletRequest(const string carrierName, const json gpslocation) {
        json findCloudletRequest;

        findCloudletRequest["ver"] = 1;
        findCloudletRequest["SessionCookie"] = sessioncookie;
        findCloudletRequest["CarrierName"] = carrierName;
        findCloudletRequest["GpsLocation"] = gpslocation;
        return findCloudletRequest;
    }

    json postRequest(const string &uri, const string &request, long &httpResponse,
            string &responseData, size_t (*responseCallback)(void *ptr, size_t size, size_t nmemb, void *s)) {
        CURL *curl;
        CURLcode res;

        cout << "URI to post to: [" << uri << "]" << endl;

        curl = curl_easy_init();
        if (curl) {
            curl_easy_setopt(curl, CURLOPT_URL, uri.c_str());
            curl_easy_setopt(curl, CURLOPT_POSTFIELDS, request.c_str());

            struct curl_slist *headers = NULL;
            headers = curl_slist_append(headers, "Accept: application/json");
            headers = curl_slist_append(headers, "Content-Type: application/json");
            headers = curl_slist_append(headers, "Charsets: utf-8");

            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
            curl_easy_setopt(curl, CURLOPT_TIMEOUT, timeoutSec);
            curl_easy_setopt(curl, CURLOPT_WRITEDATA, &responseData);
            curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, responseCallback);

            // SSL Setup:
            curl_easy_setopt(curl, CURLOPT_SSLCERT, clientCrtFile.c_str());
            curl_easy_setopt(curl, CURLOPT_SSLKEY, clientKeyFile.c_str());
            // CA:
            curl_easy_setopt(curl, CURLOPT_CAINFO, caCrtFile.c_str());

            // verify peer or disconnect
            curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);

            res = curl_easy_perform(curl);

            curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpResponse);
            if (res != CURLE_OK) {
                cout << "curl_easy_perform() failed: " << curl_easy_strerror(res) << endl;
                curl_easy_cleanup(curl);
            }

        }

        json replyData = json::parse(responseData);
        return replyData;
    }

    static size_t getReplyCallback(void *contentptr, size_t size, size_t nmemb, void *replyBuf) {
        size_t dataSize = size * nmemb;
        string *buf = ((string*)replyBuf);

        if (contentptr != NULL && buf) {
            buf->append((char*)contentptr, dataSize);
            cout << "Data Size: " << dataSize << endl;
            //cout << "Current replyBuf: [" << *buf << "]" << endl;
        }

        return dataSize;
    }

    json RegisterClient(const string &baseuri, const json &request, string &reply, long &httpResponse) {
        json jreply = postRequest(baseuri + registerAPI, request.dump(), httpResponse, reply, getReplyCallback);
        if (httpResponse != 200) {
            return jreply;
        }
        tokenserveruri = jreply["TokenServerURI"];
        sessioncookie = jreply["SessionCookie"];

        return jreply;
    }

    // string formatted json args and reply.
    json VerifyLocation(const string &baseuri, const json &request, string &reply, long &httpResponse) {
        if (tokenserveruri.size() == 0) {
            cerr << "TokenURI is empty!" << endl;
            json empty;
            return empty;
        }

        string token = getToken(tokenserveruri);
        cout << "VerifyLocation: Retrieved token: [" << token << "]" << endl;

        // Update request with the new token:
        json tokenizedRequest;
        tokenizedRequest["ver"] = request["ver"];
        tokenizedRequest["SessionCookie"] = request["SessionCookie"];
        tokenizedRequest["CarrierName"] = request["CarrierName"];
        tokenizedRequest["GpsLocation"] = request["GpsLocation"];
        tokenizedRequest["VerifyLocToken"] = token;

        json jreply = postRequest(baseuri + verifylocationAPI, tokenizedRequest.dump(), httpResponse, reply, getReplyCallback);
        return jreply;
    }

    json FindCloudlet(const string &baseuri, const json &request, string &reply, long &httpResponse) {
        json jreply = postRequest(baseuri + findcloudletAPI, request.dump(), httpResponse, reply, getReplyCallback);

        return jreply;
    }

    string getToken(const string &uri) {
        cout << "In Get Token" << endl;
        if (uri.length() == 0) {
            cerr << "No URI to get token!" << endl;
            return NULL;
        }

        CURL *curl = curl_easy_init();
        if (curl == NULL) {
            cerr << "Curl could not be initialized." << endl;
            return NULL;
        }
        CURLcode res;
        cout << "uri: " << uri << endl;
        curl_easy_setopt(curl, CURLOPT_URL, uri.c_str());
        curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, false);  // Do not follow redirect.
        curl_easy_setopt(curl, CURLOPT_HEADER, 1);              // Keep headers.

        // Set return pointer (the token), for the header callback.
        curl_easy_setopt(curl, CURLOPT_HEADERDATA, &(this->token));
        curl_easy_setopt(curl, CURLOPT_HEADERFUNCTION, token_header_callback);

        // SSL Setup:
        curl_easy_setopt(curl, CURLOPT_SSLCERT, clientCrtFile.c_str());
        curl_easy_setopt(curl, CURLOPT_SSLKEY, clientKeyFile.c_str());
        // CA:
        curl_easy_setopt(curl, CURLOPT_CAINFO, caCrtFile.c_str());
        // verify peer or disconnect
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);

        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
           cerr << "Error getting token: " << res << endl;
        }

        curl_easy_cleanup(curl);

        return token;
    }

  private:
    string token;  // short lived carrier dt-id token.
    string tokenserveruri;
    string sessioncookie;

    static string parseParameter(const string &queryParameter, const string keyFind) {
        string value;
        string foundToken;
        size_t vpos = queryParameter.find("=");

        string key = queryParameter.substr(0, vpos);
        cout << "Key: " << key << endl;
        vpos += 1; // skip over '='
        string valPart = queryParameter.substr(vpos, queryParameter.length() - vpos);
        cout << "ValPart: " << valPart << endl;
        if ((key == keyFind) && (vpos != std::string::npos)) {

            if (vpos < queryParameter.length()) {
                foundToken = queryParameter.substr(vpos, queryParameter.length() - vpos);
                cout << "Found Token: " << foundToken << endl;
            }
        }
        return foundToken;
    }

    static string parseToken(const string &locationUri) {
        // Looking for carrier dt-id: <token> in the URL, and ignoring everything else.
        size_t pos = locationUri.find("?");
        pos += 1;
        string uriParameters = locationUri.substr(pos, locationUri.length() - pos);
        pos = 0;
        size_t start = 0;
        string foundToken;

        // Carrier token.
        string keyFind = "dt-id";

        string queryParameter;
        do {
            pos = uriParameters.find("&", start);
            if (pos+1 >= uriParameters.length()) {
                break;
            }

            if (pos == std::string::npos) {  // Last one, or no terminating &
                queryParameter = uriParameters.substr(start, uriParameters.length() - start);
                foundToken = parseParameter(queryParameter, keyFind);
                break;
            } else {
                queryParameter = uriParameters.substr(start, pos - start);
                foundToken = parseParameter(queryParameter, keyFind);
            }

            // Next.
            start = pos+1;
            if (foundToken != "") {
                break;
            }
        } while (pos != std::string::npos);

        return foundToken;
    }

    static string trimStringEol(const string &stringBuf) {
        size_t size = stringBuf.length();

        // HTTP/1.1 RFC 2616: Should only be "\r\n" (and not '\n')
        if (size >= 2 && (stringBuf[size-2] == '\r' && stringBuf[size-1] == '\n')) {
            size_t seol = size-2;
            return stringBuf.substr(0,seol);
        } else {
            // Not expected EOL format, returning as-is.
            return stringBuf;
        }

    }

    // Callback function to retrieve headers, called once per header line.
    static size_t token_header_callback(const char *buffer, size_t size, size_t n_items, void *userdata) {
        size_t bufferLen = n_items * size;

        // Need to get "Location: ....dt-id=ABCDEF01234"
        string stringBuf(buffer);
        stringBuf = trimStringEol(stringBuf);

        string key = "";
        string value = "";
        string *token = (string *)userdata;

        // split buffer:
        size_t colonPos = stringBuf.find(":");
        size_t blankPos;

        if (colonPos != std::string::npos) {
            key = stringBuf.substr(0, colonPos);
            if (key == "Location") {
                // Skip blank
                blankPos = stringBuf.find(" ") + 1;
                if ((blankPos != std::string::npos) &&
                    (blankPos < stringBuf.length())) {
                    value = stringBuf.substr(blankPos, stringBuf.length() - blankPos);
                    cout << "Location Header Value: [" << value << "]" << endl;
                    *token = parseToken(value);
                }
            }
        }

        // Return number of bytes read thus far from reply stream.
        return bufferLen;
    }
};

int main() {
    cout << "Hello C++ MEX REST Sample Client" << endl;
    curl_global_init(CURL_GLOBAL_DEFAULT);

    unique_ptr<MexRestClient> mexClient = unique_ptr<MexRestClient>(new MexRestClient());

    try {
        long httpResponse;
        string baseuri;
        json loc = mexClient->retrieveLocation();

        string yn;
        cout << "Use the demo server? [yN]" << endl;
        cin >> yn;
        if (yn.compare("y") == 0) {
          baseuri = mexClient->generateBaseUri("mexdemo", mexClient->dmePort);
        } else {
          baseuri = mexClient->generateBaseUri(mexClient->getCarrierName(), mexClient->dmePort);
        }


        cout << "Register MEX client." << endl;
        cout << "====================" << endl
             << endl;


        string strRegisterClientReply;
        json registerClientRequest = mexClient->createRegisterClientRequest();
        json registerClientReply = mexClient->RegisterClient(baseuri, registerClientRequest, strRegisterClientReply, httpResponse);

        cout << "REST http response code: " << httpResponse << endl;
        if (registerClientReply.size() == 0) {
            cerr << "REST RegisterClient Error: NO RESPONSE." << endl;
            return 1;
        } else {
            cout << "REST RegisterClient Status: "
                 << ", Version: " << registerClientReply["Ver"]
                 << ", Client Status: " << registerClientReply["Status"]
                 << endl
                 << endl;
        }

        // Get the token (and wait for it)
        // GPRC uses "Channel". But, we can use libcurl here.
        cout << "Token Server URI: " << registerClientReply["TokenServerURI"] << endl;

        // Produces a new request. Now with sessioncooke and token initialized.
        cout << "Verify Location of this Mex client." << endl;
        cout << "===================================" << endl
             << endl;

        loc = mexClient->retrieveLocation();
        string strVerifyLocationReply;
        json verifyLocationRequest = mexClient->createVerifyLocationRequest(mexClient->getCarrierName(), loc, "");
        json verifyLocationReply = mexClient->VerifyLocation(baseuri, verifyLocationRequest, strVerifyLocationReply, httpResponse);

        // Print some reply values out:
        cout << "REST http response code: " << httpResponse << endl;
        if (verifyLocationReply.size() == 0) {
            cout << "REST VerifyLocation Status: NO RESPONSE" << endl;
        }
        else {
            cout << "[" << verifyLocationReply.dump() << "]" << endl;
        }

        cout << "Finding nearest Cloudlet appInsts matching this Mex client." << endl;
        cout << "===========================================================" << endl
             << endl;

        loc = mexClient->retrieveLocation();
        string strFindCloudletReply;
        json findCloudletRequest = mexClient->createFindCloudletRequest(mexClient->getCarrierName(), loc);
        json findCloudletReply = mexClient->FindCloudlet(baseuri, findCloudletRequest, strFindCloudletReply, httpResponse);

        cout << "REST http response code: " << httpResponse << endl;
        if (findCloudletReply.size() == 0) {
            cout << "REST VerifyLocation Status: NO RESPONSE" << endl;
        }
        else {
            cout << "REST FindCloudlet Status: "
                 << "Version: " << findCloudletReply["Ver"]
                 << ", Location Found Status: " << findCloudletReply["status"]
                 << ", Location of Cloudlet, Longitude: " << findCloudletReply["cloudlet_location"]["longitude"]
                 << ", Latitude: " << findCloudletReply["cloudlet_location"]["latitude"]
                 << ", Cloudlet FQDN: " << findCloudletReply["FQDN"]
                 << endl;
            json ports = findCloudletReply["ports"];
            size_t size = ports.size();
            for(const auto &appPort : ports) {
                cout << appPort.dump() << endl;
                cout << ", AppPort: Protocol: " << appPort["proto"]
                     << ", AppPort: Internal Port: " << appPort["internal_port"]
                     << ", AppPort: Public Port: " << appPort["public_port"]
                     << ", AppPort: Path Prefix: " << appPort["path_prefix"]
                     << endl;
            }
        }
        cout << "FindCloudletReply: Dump: " << findCloudletReply.dump() << endl;

        cout << endl;

    } catch (std::runtime_error &re) {
        cerr << "Runtime error occurred: " << re.what() << endl;
    } catch (std::exception &ex) {
        cerr << "Exception ocurred: " << ex.what() << endl;
    } catch (char *what) {
        cerr << "Exception: " << what << endl;
    } catch (...) {
        cerr << "Unknown failure happened." << endl;
    }

    curl_global_cleanup();
    return 0;
}
