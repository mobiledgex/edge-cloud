#include <curl/curl.h>

#include <sstream>
#include <iostream>

#include <nlohmann/json.hpp>

#include <algorithm>

#include "test_credentials.hpp"

using namespace std;
using namespace std::chrono;
using namespace nlohmann;

class MexRestClient {
  public:
    string carrierNameDefault = "tdg2";
    string baseDmeHost = "dme.mobiledgex.net";

    // API Paths:
    string registerAPI = "/v1/registerclient";
    string verifylocationAPI = "/v1/verifylocation";
    string findcloudletAPI = "/v1/findcloudlet";
    string getlocatiyonAPI = "/v1/getlocation";
    string appinstlistAPI = "/v1/getappinstlist";
    string dynamiclocgroupAPI = "/v1/dynamiclocgroup";

    unsigned long timeoutSec = 5000;
    unsigned int dmePort = 38001;
    const string appName = "EmptyMatchEngineApp"; // Your application name
    const string devName = "EmptyMatchEngineApp"; // Your developer name
    const string appVersionStr = "1.0";

    // SSL files:

    const string CaCertFile = "mex-ca.crt";
    const string ClientCertFile = "mex-client.crt";
    const string ClientPrivKey = "mex-client.key";

    MexRestClient() {}

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
        // Google's protobuf timestamp format.
        auto microseconds = std::chrono::system_clock::now().time_since_epoch();
        auto ts_sec = duration_cast<std::chrono::milliseconds>(microseconds);
        auto ts_ns = duration_cast<std::chrono::nanoseconds>(microseconds);

        json googleTimestamp;
        googleTimestamp["seconds"] = ts_sec.count();
        googleTimestamp["nanos"] = ts_sec.count();

        return googleTimestamp;
    }

    // A C++ GPS location provider/binding is needed here.
    json retrieveLocation() {
        json location;
        location["lat"] = -122.149349;
        location["long"] = 37.459609;
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

        findCloudletRequest["vers"] = 1;
        findCloudletRequest["SessionCookie"] = sessioncookie;
        findCloudletRequest["CarrierName"] = carrierName;
        findCloudletRequest["GpsLocation"] = gpslocation;
        return findCloudletRequest;
    }

    json postRequest(const string &uri, const string &request,
            string &responseData, size_t (*responseCallback)(void *ptr, size_t size, size_t nmemb, void *s)) {
        CURL *curl;
        CURLcode res;

        cout << "URI to post to: " << uri << endl;

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
            curl_easy_setopt(curl, CURLOPT_SSLCERT, ClientCertFile.c_str());
            curl_easy_setopt(curl, CURLOPT_SSLKEY, ClientPrivKey.c_str());
            // CA:
            curl_easy_setopt(curl, CURLOPT_CAINFO, CaCertFile.c_str());
            // verify peer or disconnect
            curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);

            res = curl_easy_perform(curl);
            if (res != CURLE_OK) {
                cout << "curl_easy_perform() failed: " << curl_easy_strerror(res) << endl;
                curl_easy_cleanup(curl);
            }
        }
        json replyData = json::parse(responseData);
        cout << "Reply: [" << replyData.dump() << "]" << endl;
        return replyData;
    }

    static size_t getReplyCallback(void *contentptr, size_t size, size_t nmemb, void *replyBuf) {
        size_t dataSize = size * nmemb;
        string *buf = ((string*)replyBuf);

        if (contentptr != NULL && buf) {
            string *buf = ((string*)replyBuf);
            buf->append((char*)contentptr, dataSize);

            cout << "Data Size: " << dataSize << endl;
            //cout << "Current replyBuf: [" << *buf << "]" << endl;
        }


        return dataSize;
    }

    json RegisterClient(const string &baseuri, const json &request, string &reply) {
        json jreply = postRequest(baseuri + registerAPI, request.dump(), reply, getReplyCallback);
        tokenserveruri = jreply["TokenServerURI"];
        sessioncookie = jreply["SessionCookie"];

        return jreply;
    }

    // string formatted json args and reply.
    json VerifyLocation(const string &baseuri, const json &request, string &reply) {
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

        cout << "VeriyLocation actual call..." << endl;
        json jreply = postRequest(baseuri + verifylocationAPI, tokenizedRequest.dump(), reply, getReplyCallback);
        return jreply;

    }

    json FindCloudlet(const string &baseuri, const json &request, string &reply) {
        json jreply = postRequest(baseuri + findcloudletAPI, request.dump(), reply, getReplyCallback);

        return jreply;
    }

    string getToken(const string &uri) {
        cout << "In Get Token" << endl;
        if (uri.length() == 0) {
            cerr << "No URI to get token!" << endl;
            return NULL;
        }

        // Since GPRC's Channel is somewhat hidden
        // we can use CURL here instead.
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
        curl_easy_setopt(curl, CURLOPT_HEADERFUNCTION, header_callback);

        // SSL Setup:
        curl_easy_setopt(curl, CURLOPT_SSLCERT, ClientCertFile.c_str());
        curl_easy_setopt(curl, CURLOPT_SSLKEY, ClientPrivKey.c_str());
        // CA:
        curl_easy_setopt(curl, CURLOPT_CAINFO, CaCertFile.c_str());
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
    static size_t header_callback(const char *buffer, size_t size, size_t n_items, void *userdata) {
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
    cout << "Hello C++ MEX REST Lib" << endl;
    curl_global_init(CURL_GLOBAL_DEFAULT);

    // Credentials, Mutual Authentication:
    unique_ptr<MexRestClient> mexClient = unique_ptr<MexRestClient>(new MexRestClient());

    try {
        string baseuri;
        json loc = mexClient->retrieveLocation();

        cout << "Register MEX client." << endl;
        cout << "====================" << endl
             << endl;

        baseuri = mexClient->generateBaseUri(mexClient->getCarrierName(), mexClient->dmePort);
        string strRegisterClientReply;
        json registerClientRequest = mexClient->createRegisterClientRequest();
        json registerClientReply = mexClient->RegisterClient(baseuri, registerClientRequest, strRegisterClientReply);

        if (registerClientReply.size() == 0) {
            cerr << "REST RegisterClient Error: NO RESPONSE." << endl;
            return 1;
        } else {
            cout << "REST RegisterClient Status: "
                 << "Version: " << registerClientReply["ver"]
                 << ", Client Status: " << registerClientReply["status"]
                 << ", SessionCookie: " << registerClientReply["SessionCookie"]
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

        baseuri = mexClient->generateBaseUri(mexClient->getCarrierName(), mexClient->dmePort);
        loc = mexClient->retrieveLocation();
        string strVerifyLocationReply;
        json verifyLocationRequest = mexClient->createVerifyLocationRequest(mexClient->getCarrierName(), loc, "");
        json verifyLocationReply = mexClient->VerifyLocation(baseuri, verifyLocationRequest, strVerifyLocationReply);

        // Print some reply values out:
        if (verifyLocationReply.size() == 0) {
            cout << "REST VerifyLocation Status: NO RESPONSE" << endl;
        }
        else {
            cout << "[" << verifyLocationReply.dump() << "]" << endl;
        }

        cout << "Finding nearest Cloudlet appInsts matching this Mex client." << endl;
        cout << "===========================================================" << endl
             << endl;

        baseuri = mexClient->generateBaseUri(mexClient->getCarrierName(), mexClient->dmePort);
        loc = mexClient->retrieveLocation();
        string strFindCloudletReply;
        json findCloudletRequest = mexClient->createFindCloudletRequest(mexClient->getCarrierName(), loc);
        json findCloudletReply = mexClient->FindCloudlet(baseuri, findCloudletRequest, strFindCloudletReply);

        if (findCloudletReply.size() == 0) {
            cout << "REST VerifyLocation Status: NO RESPONSE" << endl;
        }
        else {
            cout << "REST FindCloudlet Status: "
                 << "Version: " << findCloudletReply["ver"]
                 << ", Location Found Status: " << findCloudletReply["status"]
                 << ", Location of cloudlet. Latitude: " << findCloudletReply["cloudlet_location"]["lat"]
                 << ", Longitude: " << findCloudletReply["cloudlet_location"]["long"]
                 << ", Cloudlet FQDN: " << findCloudletReply["fqdn"] << endl;
            json ports = findCloudletReply["ports"];
            size_t size = ports.size();
            for(const auto &appPort : ports) {
                cout << ", AppPort: Protocol: " << appPort["proto"]
                     << ", AppPort: Internal Port: " << appPort["internal_port"]
                     << ", AppPort: Public Port: " << appPort["public_port"]
                     << ", AppPort: Public Path: " << appPort["public_path"]
                     << endl;
            }
        }

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
