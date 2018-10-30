// TEST ONLY certificates.
#include <string>

class test_credentials {
public:
    // Root CA Cert/Chian
    std::string caCrt;

    // Client Certificate
    std::string clientCrt;

    // Client Private Key:
    std::string clientKey;

    test_credentials();
};