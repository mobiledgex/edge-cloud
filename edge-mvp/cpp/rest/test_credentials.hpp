// TEST ONLY certificates.
#include <string>
#include <fstream>
#include <iostream>

using namespace std;

class test_credentials {
public:
    // Root CA Cert/Chian
    std::string caCrt;

    // Client Certificate
    std::string clientCrt;

    // Client Private Key:
    std::string clientKey;

    test_credentials(const string caCertFile, const string clientCertFile, const string clientKeyFile);

    void readPemFile(const string &filePath, string &out);
};
