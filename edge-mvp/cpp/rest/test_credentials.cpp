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

        cout << "Read for: " << filePath.c_str() << endl;
        char *buf = new char[length+1];
        ifs.read(buf, length);
        cout << "read." << endl;
        buf[length] = 0;
        cout << "terminated" << endl;

        if (!ifs) {
            cerr << "Error reading file. What's read: " << ifs.gcount() << endl;
            return out.clear();
        }
        ifs.close();

        out = buf;
        delete[] buf;
        cout << "Read [" << out << "]" << endl;
    }
}

test_credentials::test_credentials(const string caCertFile, const string clientCertFile, const string clientKeyFile) {
   readPemFile(caCertFile, this->caCrt);
   readPemFile(clientCertFile, this->clientCrt);
   readPemFile(clientKeyFile, this->clientKey);
}
