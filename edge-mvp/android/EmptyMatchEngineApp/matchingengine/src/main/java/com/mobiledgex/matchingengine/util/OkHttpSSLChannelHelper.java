package com.mobiledgex.matchingengine.util;

import android.util.Base64;

import com.mobiledgex.matchingengine.MexKeyStoreException;
import com.mobiledgex.matchingengine.MexTrustStoreException;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.security.KeyFactory;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.UnrecoverableEntryException;
import java.security.cert.Certificate;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.PKCS8EncodedKeySpec;

import javax.net.ssl.KeyManager;
import javax.net.ssl.KeyManagerFactory;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManager;
import javax.net.ssl.TrustManagerFactory;
import javax.security.auth.x500.X500Principal;

public class OkHttpSSLChannelHelper {

    public static SSLSocketFactory getMutualAuthSSLSocketFactory(String serverCAPemFileName,
                                                                 String clientCertFileName,
                                                                 String clientKeyFileName)
            throws MexKeyStoreException, MexTrustStoreException, NoSuchAlgorithmException, KeyManagementException {


        KeyManager[] keyManagers;
        TrustManager[] trustManagers;
        try {
            keyManagers = OkHttpSSLChannelHelper.getKeyManagers(clientCertFileName, clientKeyFileName);
        } catch (Exception e) {
            throw new MexKeyStoreException("Could not get Keystore: ", e);
        }

        try {
            trustManagers = OkHttpSSLChannelHelper.getTrustManagers(serverCAPemFileName);
        } catch (Exception e) {
            throw new MexTrustStoreException("Could not get Truststore: ", e);
        }

        SSLContext sslContext = SSLContext.getInstance("TLS");
        sslContext.init(
                keyManagers,
                trustManagers,
                null
                );
        return sslContext.getSocketFactory();
    }

    public static KeyManager[] getKeyManagers(String clientCertFilename, String clientPrivateKeyFilename)
            throws CertificateException, KeyStoreException, InvalidKeySpecException, IOException,
                    NoSuchAlgorithmException, UnrecoverableEntryException {

        FileInputStream clientCertStream = null;
        FileInputStream clientPrivateKeyStream = null;
        KeyManagerFactory keyManagerFactory;
        try {
            clientCertStream = new FileInputStream(clientCertFilename);
            clientPrivateKeyStream = new FileInputStream(clientPrivateKeyFilename);
            File keyFile = new File(clientPrivateKeyFilename);

            KeyStore keyStore = KeyStore.getInstance("AndroidKeyStore");
            keyStore.load(null, null);

            X500Principal principal;

            // Public Cert
            CertificateFactory cf = CertificateFactory.getInstance("X.509");
            X509Certificate pubCert = (X509Certificate) cf.generateCertificate(clientCertStream);
            principal = pubCert.getSubjectX500Principal();
            if (keyStore.getCertificate(principal.getName()) != null) {
                keyStore.deleteEntry(principal.getName("RFC2253"));
            }
            keyStore.setCertificateEntry(principal.getName("RFC2253"), pubCert);

            // Private Key for Public Cert:
            int length = (int) keyFile.length(); // byte[] can't be size long.
            if (length != keyFile.length()) {
                throw new KeyStoreException("Private Key file is too large!");
            }
            byte[] keyBytes = new byte[length];
            int read = clientPrivateKeyStream.read(keyBytes);
            if (read != keyBytes.length) {
                throw new IOException("Invalid key length read from stream.");
            }
            String keyStr = new String(keyBytes);
            keyStr = keyStr.replace("-----BEGIN RSA PRIVATE KEY-----\n", "");
            keyStr = keyStr.replace("-----END RSA PRIVATE KEY-----", "");
            byte[] decodedKey = Base64.decode(keyStr, Base64.DEFAULT);

            PrivateKey privateKey =
                    KeyFactory.getInstance("RSA").generatePrivate(new PKCS8EncodedKeySpec(decodedKey));
            keyStore.setKeyEntry(principal.getName("RFC2253"), privateKey, null, new Certificate[]{pubCert});

            // KeyManagerFactory
            keyManagerFactory = KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm());
            keyManagerFactory.init(keyStore, null);

        } finally {
            if (clientCertStream != null) {
                clientCertStream.close();
            }
            if (clientPrivateKeyStream != null) {
                clientPrivateKeyStream.close();
            }
        }
        return keyManagerFactory.getKeyManagers();
    }

    public static TrustManager[] getTrustManagers(String trustedCaFilename)
            throws CertificateException, IOException, KeyStoreException, NoSuchAlgorithmException {

        FileInputStream caCertFileStream = null;
        TrustManagerFactory trustManagerFactory = null;
        try {
            // For TrustCA.
            KeyStore keyStore = KeyStore.getInstance(KeyStore.getDefaultType());
            keyStore.load(null);
            caCertFileStream = new FileInputStream(new File(trustedCaFilename));

            CertificateFactory cf = CertificateFactory.getInstance("X.509");
            X509Certificate caCert = (X509Certificate) cf.generateCertificate(caCertFileStream);

            X500Principal principal = caCert.getSubjectX500Principal();
            keyStore.setCertificateEntry(principal.getName("RFC2253"), caCert);

            // Set up trust manager factory to use our key store.
            trustManagerFactory = TrustManagerFactory.getInstance(
                    TrustManagerFactory.getDefaultAlgorithm());
            trustManagerFactory.init(keyStore);
        } finally {
            if (caCertFileStream != null) {
                caCertFileStream.close();
            }
        }

        return trustManagerFactory.getTrustManagers();
    }
}
