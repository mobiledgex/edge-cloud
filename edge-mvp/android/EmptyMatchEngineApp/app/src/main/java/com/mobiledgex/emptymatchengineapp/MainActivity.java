package com.mobiledgex.emptymatchengineapp;

import android.app.Activity;
import android.content.SharedPreferences;
import android.content.res.Resources;
import android.location.Location;
import android.os.AsyncTask;
import android.os.Bundle;
import android.preference.PreferenceManager;
import android.support.design.widget.FloatingActionButton;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.Toolbar;
import android.telephony.CarrierConfigManager;
import android.util.Log;
import android.view.View;
import android.view.Menu;
import android.view.MenuItem;
import android.widget.TextView;

import android.content.Intent;

// Matching Engine API:
import com.mobiledgex.matchingengine.MatchingEngine;
import com.mobiledgex.matchingengine.NetworkRequestTimeoutException;
import com.mobiledgex.matchingengine.util.RequestPermissions;

import distributed_match_engine.AppClient;
import distributed_match_engine.Appcommon;
import io.grpc.StatusRuntimeException;


// Location API:
import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.android.gms.location.LocationServices;
import com.google.android.gms.location.LocationCallback;
import com.google.android.gms.location.LocationRequest;
import com.google.android.gms.location.LocationResult;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;

import java.io.IOException;
import java.util.List;
import java.util.Locale;
import java.util.concurrent.ExecutionException;


public class MainActivity extends AppCompatActivity implements SharedPreferences.OnSharedPreferenceChangeListener {
    private static final String TAG = "MainActivity";
    private MatchingEngine mMatchingEngine;
    private String someText = null;

    private RequestPermissions mRpUtil;
    private FusedLocationProviderClient mFusedLocationClient;

    private LocationCallback mLocationCallback;
    private LocationRequest mLocationRequest;
    private boolean mDoLocationUpdates;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        /**
         * MatchEngine APIs require special user approved permissions to READ_PHONE_STATE and
         * one of the following:
         * ACCESS_FINE_LOCATION or ACCESS_COARSE_LOCATION. This creates a dialog, if needed.
         */
        mRpUtil = new RequestPermissions();
        mFusedLocationClient = LocationServices.getFusedLocationProviderClient(this);
        mLocationRequest = new LocationRequest();

        mMatchingEngine = new MatchingEngine(this);

        // Restore mex location preference, defaulting to false:
        SharedPreferences prefs = PreferenceManager.getDefaultSharedPreferences(this);
        PreferenceManager.setDefaultValues(this, R.xml.location_preferences, false);

        boolean mexLocationAllowed = prefs.getBoolean(getResources()
                .getString(R.string.preference_mex_location_verification),
                false);
        MatchingEngine.setMexLocationAllowed(mexLocationAllowed);

        // Watch allowed preference:
        prefs.registerOnSharedPreferenceChangeListener(this);


        // Client side FusedLocation updates.
        mDoLocationUpdates = true;

        setContentView(R.layout.activity_main);
        Toolbar toolbar = findViewById(R.id.toolbar);
        setSupportActionBar(toolbar);

        mLocationCallback = new LocationCallback() {
            @Override
            public void onLocationResult(LocationResult locationResult) {
                if (locationResult == null) {
                    return;
                }
                String clientLocText = "";
                for (Location location : locationResult.getLocations()) {
                    // Update UI with client location data
                    clientLocText += "[" + location.toString() + "]";
                }
                TextView tv = findViewById(R.id.client_location_content);
                tv.setText(clientLocText);
            };
        };

        FloatingActionButton fab = findViewById(R.id.fab);
        fab.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                doEnhancedLocationVerification();
            }
        });

        Toolbar myToolbar = findViewById(R.id.toolbar);
        setSupportActionBar(myToolbar);

        // Open dialog for MEX if this is the first time the app is created:

        String firstTimeUsePrefKey = getResources().getString(R.string.preference_first_time_use);
        boolean firstTimeUse = prefs.getBoolean(firstTimeUsePrefKey, true);

        if (firstTimeUse) {
            Intent intent = new Intent(this, FirstTimeUseActivity.class);
            startActivity(intent);
        }
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.menu_main, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();

        if (id == R.id.action_settings) {

            // Open "Settings" UI
            Intent intent = new Intent(this, SettingsActivity.class);
            startActivity(intent);

            return true;
        }

        return super.onOptionsItemSelected(item);
    }

    @Override
    public void onResume() {
        super.onResume();

        if (mDoLocationUpdates) {
            startLocationUpdates();
        }
    }

    @Override
    public void onPause() {
        super.onPause();
        stopLocationUpdates();
    }

    @Override
    public void onSaveInstanceState(Bundle savedInstanceState) {
        super.onSaveInstanceState(savedInstanceState);
        savedInstanceState.putBoolean(MatchingEngine.MEX_LOCATION_PERMISSION, MatchingEngine.isMexLocationAllowed());
    }

    @Override
    public void onRestoreInstanceState(Bundle restoreInstanceState) {
        super.onRestoreInstanceState(restoreInstanceState);
        if (restoreInstanceState != null) {
            MatchingEngine.setMexLocationAllowed(restoreInstanceState.getBoolean(MatchingEngine.MEX_LOCATION_PERMISSION));
        }
    }

    /**
     * See documentation for Google's FusedLocationProviderClient for additional usage information.
     */
    private void startLocationUpdates() {
        // As of Android 23, permissions can be asked for while the app is still running.
        if (mRpUtil.getNeededPermissions(this).size() > 0) {
            return;
        }

        try {
            if (mFusedLocationClient == null) {
                mFusedLocationClient = LocationServices.getFusedLocationProviderClient(this);
            }
            mFusedLocationClient.requestLocationUpdates(mLocationRequest,mLocationCallback,
                    null /* Looper */);
        } catch (SecurityException se) {
            se.printStackTrace();
            Log.i(TAG, "App should Request location permissions during onCreate().");
        }
    }

    private void stopLocationUpdates() {
        mFusedLocationClient.removeLocationUpdates(mLocationCallback);
    }

    @Override
    public void onSharedPreferenceChanged(SharedPreferences sharedPreferences, String key) {
        final String prefKeyAllowMEX = getResources().getString(R.string.preference_mex_location_verification);

        if (key.equals(prefKeyAllowMEX)) {
            boolean mexLocationAllowed = sharedPreferences.getBoolean(prefKeyAllowMEX, false);
            MatchingEngine.setMexLocationAllowed(mexLocationAllowed);
        }
    }

    public void doEnhancedLocationVerification() throws SecurityException {
        final Activity ctx = this;

        // As of Android 23, permissions can be asked for while the app is still running.
        if (mRpUtil.getNeededPermissions(this).size() > 0) {
            mRpUtil.requestMultiplePermissions(this);
            return;
        }

        // Run in the background and post text results to the UI thread.
        mFusedLocationClient.getLastLocation().addOnCompleteListener(new OnCompleteListener<Location>() {
            @Override
            public void onComplete(Task<Location> task) {
                if (task.isSuccessful() && task.getResult() != null) {
                    doEnhancedLocationUpdateInBackground(task, ctx);
                } else {
                    Log.w(TAG, "getLastLocation:exception", task.getException());
                    someText = "Last location not found, or has never been used. Location cannot be verified using 'getLastLocation()'. " +
                            "Use the requestLocationUpdates() instead if applicable for location verification.";
                    TextView tv = findViewById(R.id.mex_verified_location_content);
                    tv.setText(someText);
                }
            }
        });
    }

    private void doEnhancedLocationUpdateInBackground(final Task<Location> aTask, final Activity ctx) {
        AsyncTask.execute(new Runnable() {
            @Override
            public void run() {
                Location location = aTask.getResult();
                // Location found. Create a request:

                try {
                    SharedPreferences prefs = PreferenceManager.getDefaultSharedPreferences(ctx);
                    boolean mexAllowed = prefs.getBoolean(getResources().getString(R.string.preference_mex_location_verification), false);

                    //String carrierName = mMatchingEngine.retrieveNetworkCarrierName(ctx); // Regular use case
                    String carrierName = "mexdemo";                                         // Override carrierName
                    //String carrierName = mMatchingEngine.retrieveNetworkCarrierName(ctx); // Regular use case
                    if (carrierName == null) {
                        someText = "No carrier Info!";
                    }
                    String host = mMatchingEngine.generateDmeHostAddress(carrierName);      // Override carrier specific host name
                    int port = mMatchingEngine.getPort(); // Keep same port.

                    String devName = "EmptyMatchEngineApp"; // Always supplied by application.

                    AppClient.RegisterClientRequest registerClientRequest =
                            mMatchingEngine.createRegisterClientRequest(ctx,
                                    devName, mMatchingEngine.getAppName(ctx),
                                    null, carrierName, null);

                    AppClient.RegisterClientReply registerStatus =
                            mMatchingEngine.registerClient(registerClientRequest,
                                    host, port, 10000);

                    if (registerStatus.getStatus() != AppClient.ReplyStatus.RS_SUCCESS) {
                        someText = "Registration Failed. Error: " + registerStatus.getStatus();
                        TextView tv = findViewById(R.id.mex_verified_location_content);
                        tv.setText(someText);
                        return;
                    }

                    AppClient.VerifyLocationRequest verifyRequest =
                            mMatchingEngine.createVerifyLocationRequest(ctx, carrierName, location);
                    if (verifyRequest != null) {
                        // Location Verification (Blocking, or use verifyLocationFuture):
                        AppClient.VerifyLocationReply verifiedLocation =
                                mMatchingEngine.verifyLocation(verifyRequest, host, port, 10000);

                        someText = "[Location Verified: Tower: " + verifiedLocation.getTowerStatus() +
                                ", GPS LocationStatus: " + verifiedLocation.getGpsLocationStatus() +
                                ", Location Accuracy: " + verifiedLocation.getGPSLocationAccuracyKM() + " ]\n";

                        // Find the closest cloudlet for your application to use. (Blocking call, or use findCloudletFuture):
                        AppClient.FindCloudletRequest findCloudletRequest =
                                mMatchingEngine.createFindCloudletRequest(ctx, carrierName, location);
                        AppClient.FindCloudletReply closestCloudlet = mMatchingEngine.findCloudlet(findCloudletRequest,
                                host, port, 10000);

                        List<distributed_match_engine.Appcommon.AppPort> ports = closestCloudlet.getPortsList();
                        String portListStr = "";
                        boolean first = true;
                        String appPortFormat = "{Protocol: %d, Container Port: %d, External Port: %d, Path Prefix: '%s'}";
                        for (Appcommon.AppPort aPort : ports) {
                            if (!first) {
                                portListStr += ", ";

                            }
                            portListStr += String.format(Locale.getDefault(), appPortFormat,
                                aPort.getProto().getNumber(),
                                aPort.getInternalPort(),
                                aPort.getPublicPort(),
                                aPort.getPathPrefix());
                        }

                        someText += "[Cloudlet App Ports: [" + portListStr + "]\n";

                        String appInstListText = "";
                        AppClient.AppInstListRequest appInstListRequest = mMatchingEngine.createAppInstListRequest(ctx, carrierName, location);
                        AppClient.AppInstListReply appInstListReply = mMatchingEngine.getAppInstList(appInstListRequest,10000);
                        for (AppClient.CloudletLocation cloudletLocation : appInstListReply.getCloudletsList()) {
                            String location_carrierName = cloudletLocation.getCarrierName();
                            String location_cloudletName = cloudletLocation.getCloudletName();
                            double location_distance = cloudletLocation.getDistance();

                            appInstListText += "[CloudletLocation: CarrierName: " + location_carrierName;
                            appInstListText += ", CloudletName: " + location_cloudletName;
                            appInstListText += ", Distance: " + location_distance;
                            appInstListText += " , AppInstances: [";
                            for (AppClient.Appinstance appinstance : cloudletLocation.getAppinstancesList()) {
                                appInstListText += "Name: " + appinstance.getAppName()
                                                + ", Version: " + appinstance.getAppVers()
                                                + ", FQDN: " + appinstance.getFQDN()
                                                + ", Ports: " + appinstance.getPortsList().toString();

                            }
                            appInstListText += "]]";
                        }
                        if (!appInstListText.isEmpty()) {
                            someText += appInstListText;
                        }
                    } else {
                        someText = "Cannot create request object.\n";
                        if (!mexAllowed) {
                            someText += " Reason: Enhanced location is disabled.\n";
                        }
                    }

                    someText += "[Is WiFi Enabled: " + mMatchingEngine.isWiFiEnabled(ctx) + "]\n";

                    if (android.os.Build.VERSION.SDK_INT >= 28) {
                        someText += "[Is Roaming Data Enabled: " + mMatchingEngine.isRoamingData() + "]\n";
                    } else {
                        someText += "[Roaming Data status unknown.]\n";
                    }

                    CarrierConfigManager carrierConfigManager = ctx.getSystemService(CarrierConfigManager.class);
                    someText += "[Enabling WiFi Calling could disable Cellular Data if on a Roaming Network!\nWiFi Calling  Support Status: "
                            + mMatchingEngine.isWiFiCallingSupported(carrierConfigManager) + "]\n";


                    // Background thread. Post update to the UI thread:
                    ctx.runOnUiThread(new Runnable() {
                        @Override
                        public void run() {
                            TextView tv = findViewById(R.id.mex_verified_location_content);
                            tv.setText(someText);
                        }
                    });
                } catch (ExecutionException ee) {
                    ee.printStackTrace();
                    if (ee.getCause() instanceof NetworkRequestTimeoutException) {
                        String causeMessage = ee.getCause().getMessage();
                        someText = "Network connection failed: " + causeMessage;
                        Log.e(TAG, someText);
                        // Handle network error with failover logic. MEX requests over cellular is needed to talk to the DME.
                        ctx.runOnUiThread(new Runnable() {
                            @Override
                            public void run() {
                                TextView tv = findViewById(R.id.mex_verified_location_content);
                                tv.setText(someText);
                            }
                        });
                    }
                } catch (IOException ioe) {
                    ioe.printStackTrace();
                } catch (InterruptedException ie) {
                    ie.printStackTrace();
                } catch (StatusRuntimeException sre) {
                    sre.printStackTrace();
                } catch (IllegalArgumentException iae) {
                    iae.printStackTrace();
                } catch (Resources.NotFoundException nfe) {
                    nfe.printStackTrace();
                }
            }
        });
    }
}
