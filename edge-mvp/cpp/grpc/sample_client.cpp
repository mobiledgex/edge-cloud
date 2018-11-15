#include <grpcpp/grpcpp.h>

#include <iostream>

#include <curl/curl.h>

#include "app-client.grpc.pb.h"
#include "test_credentials.hpp"

using namespace std;
using namespace std::chrono;
using namespace distributed_match_engine;
using distributed_match_engine::Match_Engine_Api;

using grpc::Channel;
using grpc::ClientContext;
using grpc::Status;


// Test Cert files:
struct MutualAuthFiles {
    const string caCrtFile = "../../../tls/out/mex-ca.crt";
    const string clientCrtFile = "../../../tls/out/mex-client.crt";
    const string clientKeyFile = "../../../tls/out/mex-client.key";
} mutualAuthFiles;

class MexGrpcClient {
  public:
    unsigned long timeoutSec = 5000;
    const string appName = "EmptyMatchEngineApp"; // Your application name
    const string devName = "EmptyMatchEngineApp"; // Your developer name
    const string appVersionStr = "1.0";

    MexGrpcClient(std::shared_ptr<Channel> channel)
        : stub_(Match_Engine_Api::NewStub(channel)) {}

    // Retrieve the carrier name of the cellular network interface.
    string getCarrierName() {
        return string("TDG");
    }

    // A C++ GPS location provider/binding is needed here.
    //unique_ptr<Loc> location = // make_unique --> C++14...
    unique_ptr<Loc> retrieveLocation() {
        auto location = unique_ptr<Loc>(new Loc());
        if (location == NULL) {
            throw "Location allocation failed!";
        }
        location->set_long_(-122.149349);
        location->set_lat(37.459609);
        location->set_horizontal_accuracy(5);
        location->set_vertical_accuracy(20);
        location->set_altitude(100);
        location->set_course(0);
        location->set_speed(2);

        // Google's protobuf timestamp
        auto timestampPtr = new Timestamp();
        auto ts_micro = std::chrono::system_clock::now().time_since_epoch();
        auto ts_sec = duration_cast<std::chrono::seconds>(ts_micro);
        auto ts_micro_remainder = ts_micro % std::chrono::microseconds(1000000);
        auto ts_nano_remainder = duration_cast<std::chrono::nanoseconds>(ts_micro_remainder);

        timestampPtr->set_seconds(ts_sec.count());
        timestampPtr->set_nanos(ts_nano_remainder.count());

        location->set_allocated_timestamp(timestampPtr);
        return location;
    }

    unique_ptr<RegisterClientRequest> createRegisterClientRequest(const string &authToken) {
        unique_ptr<RegisterClientRequest> request = unique_ptr<RegisterClientRequest>(new RegisterClientRequest());

        request->set_ver(1);

        request->set_devname(devName);
        request->set_appname(appName);
        request->set_appvers(appVersionStr);
        request->set_authtoken(authToken);

        return request;
    }

    // Carrier name can change depending on cell tower.
    unique_ptr<VerifyLocationRequest> createVerifyLocationRequest(const string carrierName, const shared_ptr<Loc> gpslocation, const string verifyloctoken) {
        unique_ptr<VerifyLocationRequest> request = unique_ptr<VerifyLocationRequest>(new VerifyLocationRequest());

        request->set_ver(1);

        request->set_sessioncookie(sessioncookie);
        request->set_carriername(carrierName);

        Loc *ownedLocation = gpslocation->New();
        ownedLocation->CopyFrom(*gpslocation);
        request->set_allocated_gpslocation(ownedLocation);

        // This is a carrier supplied token.
        if (verifyloctoken.length() != 0) {
            request->set_verifyloctoken(verifyloctoken);
        }

        return request;
    }

    // Carrier name can change depending on cell tower.
    unique_ptr<FindCloudletRequest> createFindCloudletRequest(const string carrierName, const shared_ptr<Loc> gpslocation) {
        unique_ptr<FindCloudletRequest> request = unique_ptr<FindCloudletRequest>(new FindCloudletRequest());

        request->set_ver(1);

        request->set_sessioncookie(sessioncookie);
        request->set_carriername(carrierName);

        Loc *ownedLocation = gpslocation->New();
        ownedLocation->CopyFrom(*gpslocation);
        request->set_allocated_gpslocation(ownedLocation);

        return request;
    }

    grpc::Status RegisterClient(const shared_ptr<RegisterClientRequest> request, RegisterClientReply &reply) {
        // As per GRPC documenation, DO NOT REUSE contexts across RPCs.
        ClientContext context;

        // Context deadline is in seconds.
        system_clock::time_point deadline = chrono::system_clock::now() + chrono::seconds(timeoutSec);
        context.set_deadline(deadline);

        grpc::Status grpcStatus = stub_->RegisterClient(&context, *request, &reply);

        // Save some Mex state info for other calls.
        this->sessioncookie = reply.sessioncookie();
        this->tokenserveruri = reply.tokenserveruri();

        return grpcStatus;
    }

    grpc::Status VerifyLocation(const shared_ptr<VerifyLocationRequest> request, VerifyLocationReply &reply) {
        string token = getToken(tokenserveruri);

        ClientContext context;
        // Context deadline is in seconds.
        system_clock::time_point deadline = chrono::system_clock::now() + chrono::seconds(timeoutSec);
        context.set_deadline(deadline);

        // Recreate request with the new token:
        unique_ptr<VerifyLocationRequest> tokenizedRequest = unique_ptr<VerifyLocationRequest>(new VerifyLocationRequest());
        tokenizedRequest->CopyFrom(*request);

        tokenizedRequest->set_verifyloctoken(token);
        grpc::Status grpcStatus = stub_->VerifyLocation(&context, *tokenizedRequest, &reply);

        return grpcStatus;
    }

    grpc::Status FindCloudlet(const shared_ptr<FindCloudletRequest> request, FindCloudletReply &reply) {
        ClientContext context;
        // Context deadline is in seconds.
        system_clock::time_point deadline = chrono::system_clock::now() + chrono::seconds(timeoutSec);
        context.set_deadline(deadline);

        grpc::Status grpcStatus = stub_->FindCloudlet(&context, *request, &reply);

        return grpcStatus;
    }

    string getToken(const string &uri) {
        cout << "In getToken" << endl;
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
        curl_easy_setopt(curl, CURLOPT_SSLCERT, mutualAuthFiles.clientCrtFile.c_str());
        curl_easy_setopt(curl, CURLOPT_SSLKEY, mutualAuthFiles.clientKeyFile.c_str());
        // CA:
        curl_easy_setopt(curl, CURLOPT_CAINFO, mutualAuthFiles.caCrtFile.c_str());

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
    std::unique_ptr<Match_Engine_Api::Stub> stub_;
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
        string valPart = queryParameter.substr(vpos, queryParameter.length());
        if ((key == keyFind) && (vpos != std::string::npos)) {

            if (vpos < queryParameter.length()) {
                foundToken = queryParameter.substr(vpos, queryParameter.length());
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
                cout << "Substring: " << queryParameter << endl;
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
                    cout << "Location Header Value: " << value << endl;
                    *token = parseToken(value);
                }
            }
        }

        // Return number of bytes read thus far from reply stream.
        return bufferLen;
    }
};

int main() {
    cout << "Hello C++ MEX GRPC Lib" << endl;
    string host = "tdg2.dme.mobiledgex.net:50051";

    // Credentials, Mutual Authentication:
    unique_ptr<test_credentials> test_creds = unique_ptr<test_credentials>(
            new test_credentials(
                mutualAuthFiles.caCrtFile,
                mutualAuthFiles.clientCrtFile,
                mutualAuthFiles.clientKeyFile));

    grpc::SslCredentialsOptions credentials;

    credentials.pem_root_certs = test_creds->caCrt;
    credentials.pem_cert_chain = test_creds->clientCrt;
    credentials.pem_private_key = test_creds->clientKey;

    auto channel_creds = grpc::SslCredentials(grpc::SslCredentialsOptions(credentials));
    shared_ptr<Channel> channel = grpc::CreateChannel(host, channel_creds);

    unique_ptr<MexGrpcClient> mexClient = unique_ptr<MexGrpcClient>(new MexGrpcClient(channel));

    try {
        shared_ptr<Loc> loc = mexClient->retrieveLocation();

        cout << "Register MEX client." << endl;
        cout << "====================" << endl
             << endl;

        RegisterClientReply registerClientReply;
        shared_ptr<RegisterClientRequest> registerClientRequest = unique_ptr<RegisterClientRequest>(mexClient->createRegisterClientRequest(""));
        grpc::Status grpcStatus = mexClient->RegisterClient(registerClientRequest, registerClientReply);

        if (!grpcStatus.ok()) {
            cerr << "GPRC RegisterClient Error: " << grpcStatus.error_message() << endl;
            return 1;
        } else {
            cout << "GRPC RegisterClient Status: "
                 << grpcStatus.error_code()
                 << ", Version: " << registerClientReply.ver()
                 << ", Client Status: " << registerClientReply.status()
                 << endl
                 << endl;
        }

        // Get the token (and wait for it)
        // GPRC uses "Channel". But, we can use libcurl here.
        cout << "Token Server URI: " << registerClientReply.tokenserveruri() << endl;

        // Produces a new request. Now with sessioncooke and token initialized.
        cout << "Verify Location of this Mex client." << endl;
        cout << "===================================" << endl
             << endl;

        VerifyLocationReply verifyLocationReply;
        loc = mexClient->retrieveLocation();
        shared_ptr<VerifyLocationRequest> verifyLocationRequest = unique_ptr<VerifyLocationRequest>(
                    mexClient->createVerifyLocationRequest(mexClient->getCarrierName(), loc, ""));
        grpcStatus = mexClient->VerifyLocation(verifyLocationRequest, verifyLocationReply);

        if (!grpcStatus.ok()) {
            cerr << "GPRC VerifyLocation Error: " << grpcStatus.error_message() << endl;
        } else {
            // Print some reply values out:
            cout << "GRPC VerifyLocation Status: " << grpcStatus.error_code()
                 << ", Version: " << verifyLocationReply.ver()
                 << ", Tower Status: " << verifyLocationReply.tower_status()
                 << ", GPS Location Status: " << verifyLocationReply.gps_location_status()
                 << ", GPS Location Accuracy: " << verifyLocationReply.gps_location_accuracy_km()
                 << endl
                 << endl;
        }

        cout << "Finding nearest Cloudlet appInsts matching this Mex client." << endl;
        cout << "===========================================================" << endl
             << endl;

        FindCloudletReply findCloudletReply;
        loc = mexClient->retrieveLocation();
        shared_ptr<FindCloudletRequest> findCloudletRequest = unique_ptr<FindCloudletRequest>(
                    mexClient->createFindCloudletRequest(mexClient->getCarrierName(), loc));
        grpcStatus = mexClient->FindCloudlet(findCloudletRequest, findCloudletReply);

        if (!grpcStatus.ok()) {
            cerr << "GPRC FindCloudlet Error: " << grpcStatus.error_message() << endl;
        } else {
            cout << "GRPC FindCloudlet Status: " << grpcStatus.error_code()
                 << ", Version: " << findCloudletReply.ver()
                 << ", Location Found Status: " << findCloudletReply.status()
                 << ", Location of cloudlet. Latitude: " << findCloudletReply.cloudlet_location().lat()
                 << ", Longitude: " << findCloudletReply.cloudlet_location().long_()
                 << ", Cloudlet FQDN: " << findCloudletReply.fqdn() << endl;

            for (int i = 0; i < findCloudletReply.ports_size(); i++) {
                cout << ", AppPort: Protocol: " << findCloudletReply.ports().Get(i).proto()
                     << ", AppPort: Internal Port: " << findCloudletReply.ports().Get(i).internal_port()
                     << ", AppPort: Public Port: " << findCloudletReply.ports().Get(i).public_port()
                     << ", AppPort: Public Path: " << findCloudletReply.ports().Get(i).public_path()
                     << endl;
            }
            cout << endl;
        }

    } catch (std::runtime_error &re) {
        cerr << "Runtime error occurred: " << re.what() << endl;
    } catch (std::exception &ex) {
        cerr << "Exception ocurred: " << ex.what() << endl;
    } catch (char *what) {
        cerr << "Exception: " << what << endl;
    } catch (...) {
        cerr << "Unknown failure happened." << endl;
    }

    return 0;
}
