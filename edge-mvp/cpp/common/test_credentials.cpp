#include "test_credentials.hpp"

// TEST ONLY certificates.
void test_credentials::readPemFile(const string &filePath, string &out) {
    std::ifstream ifs (filePath, std::ifstream::binary);
    if (ifs) {
        ifs.seekg(0, ifs.end);
        size_t length = ifs.tellg();
        ifs.seekg(0, ifs.beg);

        cout << "File Length: " << length << endl;
        if (!length) {
            return;
        }

        cout << "Reading: " << filePath.c_str() << endl;
        char *buf = new char[length+1];
        ifs.read(buf, length);
        buf[length] = 0;

        if (!ifs) {
            cerr << "Error reading file. What's read: " << ifs.gcount() << endl;
            return out.clear();
        }
        ifs.close();

        out = buf;
        delete[] buf;
    }
}

test_credentials::test_credentials(const string caCertFile, const string clientCertFile, const string clientKeyFile) {
   readPemFile(caCertFile, this->caCrt);
   readPemFile(clientCertFile, this->clientCrt);
   readPemFile(clientKeyFile, this->clientKey);
}
