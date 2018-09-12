package com.mobiledgex.matchingengine.util;

import android.content.Context;
import android.content.res.AssetManager;
import android.util.Base64;
import android.util.Log;

import com.mobiledgex.matchingengine.MexKeyStoreException;
import com.mobiledgex.matchingengine.MexTrustStoreException;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
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

    public static SSLSocketFactory getMutualAuthSSLSocketFactory(FileInputStream serverCAPemFis,
                                                                 FileInputStream clientCertFis,
                                                                 PrivateKey privateKey)
            throws MexKeyStoreException, MexTrustStoreException, NoSuchAlgorithmException, KeyManagementException {

        KeyManager[] keyManagers;
        TrustManager[] trustManagers;
        try {
            keyManagers = OkHttpSSLChannelHelper.getKeyManagers(clientCertFis, privateKey);
        } catch (Exception e) {
            throw new MexKeyStoreException("Could not get Keystore: ", e);
        }

        try {
            trustManagers = OkHttpSSLChannelHelper.getTrustManagers(serverCAPemFis);
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

    public static PrivateKey getPrivateKey(String clientPrivateKeyFilename)
        throws IOException, InvalidKeySpecException, KeyStoreException, NoSuchAlgorithmException {
        // Private Key for Public Cert:
        PrivateKey privateKey = null;
        FileInputStream clientPrivateKeyStream = null;
        try {
            File keyFile = new File(clientPrivateKeyFilename);
            clientPrivateKeyStream = new FileInputStream(keyFile);

            int length = (int) keyFile.length(); // byte[] can't be size long.
            if (length != keyFile.length()) {
                throw new KeyStoreException("Private Key file is too large to convert to byte array!");
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
            privateKey = KeyFactory.getInstance("RSA").generatePrivate(new PKCS8EncodedKeySpec(decodedKey));
        } finally {
            if (clientPrivateKeyStream != null) {
                clientPrivateKeyStream.close();
            }
        }
        return privateKey;
    }

    public static KeyManager[] getKeyManagers(FileInputStream clientCertStream, PrivateKey privateKey)
            throws CertificateException, KeyStoreException, IOException,
                    NoSuchAlgorithmException, UnrecoverableEntryException {

        KeyManagerFactory keyManagerFactory;

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

        // KeyEntry with Private Key with public Certificate
        keyStore.setKeyEntry(principal.getName("RFC2253"), privateKey, null, new Certificate[]{pubCert});

        // KeyManagerFactory
        keyManagerFactory = KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm());
        keyManagerFactory.init(keyStore, null);

        return keyManagerFactory.getKeyManagers();
    }

    public static TrustManager[] getTrustManagers(FileInputStream caCertFileStream)
            throws CertificateException, IOException, KeyStoreException, NoSuchAlgorithmException {

        TrustManagerFactory trustManagerFactory = null;

        // For TrustCA.
        KeyStore keyStore = KeyStore.getInstance(KeyStore.getDefaultType());
        keyStore.load(null);

        CertificateFactory cf = CertificateFactory.getInstance("X.509");
        X509Certificate caCert = (X509Certificate) cf.generateCertificate(caCertFileStream);

        X500Principal principal = caCert.getSubjectX500Principal();
        keyStore.setCertificateEntry(principal.getName("RFC2253"), caCert);

        // Set up trust manager factory to use our key store.
        trustManagerFactory = TrustManagerFactory.getInstance(
                TrustManagerFactory.getDefaultAlgorithm());
        trustManagerFactory.init(keyStore);


        return trustManagerFactory.getTrustManagers();
    }

    public static void copyAssets(Context context, String sourceDir, String outputDir) {
        Log.d("copyAssets", "copyAssets(" + sourceDir + "," + outputDir + ")");
        AssetManager assetManager = context.getAssets();
        String[] files = null;
        try {
            files = assetManager.list(sourceDir);
        } catch (IOException e) {
            Log.e("copyAssets", "Failed to get asset file list.", e);
        }
        if (files != null) {
            for (String filename : files) {
                Log.d("copyAssets", "filename=" + filename);
                InputStream in = null;
                OutputStream out = null;
                try {
                    in = assetManager.open(sourceDir + "/" + filename);
                    File outFile = new File(outputDir, filename);
                    Log.d("copyAssets", "outFile=" + outFile.getAbsolutePath());
                    out = new FileOutputStream(outFile);
                    copyFile(in, out);
                } catch (IOException e) {
                    Log.e("copyAssets", "Failed to copy asset file: " + filename, e);
                } finally {
                    if (in != null) {
                        try {
                            in.close();
                        } catch (IOException e) {
                            // NOOP
                        }
                    }
                    if (out != null) {
                        try {
                            out.close();
                        } catch (IOException e) {
                            // NOOP
                        }
                    }
                }
            }
        }
    }

    public static void copyFile(InputStream in, OutputStream out) throws IOException {
        byte[] buffer = new byte[1024];
        int read;
        while((read = in.read(buffer)) != -1){
            out.write(buffer, 0, read);
        }
    }

}
