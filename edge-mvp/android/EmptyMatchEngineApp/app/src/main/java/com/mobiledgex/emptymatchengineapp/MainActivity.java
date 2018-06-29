package com.mobiledgex.emptymatchengineapp;

import android.app.Activity;
import android.content.SharedPreferences;
import android.location.Location;
import android.os.Bundle;
import android.preference.PreferenceManager;
import android.support.annotation.NonNull;
import android.support.design.widget.FloatingActionButton;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.Toolbar;
import android.util.Log;
import android.view.View;
import android.view.Menu;
import android.view.MenuItem;
import android.widget.TextView;

import android.content.Intent;

// Matching Engine API:
import com.mobiledgex.matchingengine.FindCloudletResponse;
import com.mobiledgex.matchingengine.MatchingEngine;
import com.mobiledgex.matchingengine.util.RequestPermissions;

import distributed_match_engine.AppClient;
import io.grpc.StatusRuntimeException;


// Location API:
import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.android.gms.location.LocationServices;
import com.google.android.gms.location.LocationCallback;
import com.google.android.gms.location.LocationRequest;
import com.google.android.gms.location.LocationResult;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;


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
        mRpUtil.requestMultiplePermissions(this);
        mFusedLocationClient = LocationServices.getFusedLocationProviderClient(this);
        mLocationRequest = new LocationRequest();

        mMatchingEngine = new MatchingEngine();

        // Restore mex location preference, defaulting to false:
        SharedPreferences prefs = PreferenceManager.getDefaultSharedPreferences(this);
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
        boolean firstTimeUse = prefs.getBoolean(getResources().getString(R.string.perference_first_time_use), true);
        if (firstTimeUse) {
            new EnhancedLocationDialog().show(this.getSupportFragmentManager(), "dialog");
            String firstTimeUseKey = getResources().getString(R.string.perference_first_time_use);
            // Disable first time use.
            prefs.edit()
                    .putBoolean(firstTimeUseKey, false)
                    .apply();
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

        //noinspection SimplifiableIfStatement
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
    public void onRequestPermissionsResult(int requestCode, @NonNull String[] permissions,
                                           @NonNull int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        // Or replace with an app specific dialog set.
        mRpUtil.onRequestPermissionsResult(this, requestCode, permissions, grantResults);
    }


    @Override
    public void onSharedPreferenceChanged(SharedPreferences sharedPreferences, String key) {
        final String prefKeyAllowMEX = getResources().getString(R.string.preference_mex_location_verification);

        if (key.equals(prefKeyAllowMEX)) {
            SharedPreferences prefs = PreferenceManager.getDefaultSharedPreferences(getApplicationContext());
            boolean mexLocationAllowed = prefs.getBoolean(prefKeyAllowMEX, false);
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

        // Run in the background and post to the UI thread. Here, it is simply attached to a button
        // on the current UI Thread.
        mFusedLocationClient.getLastLocation().addOnCompleteListener(new OnCompleteListener<Location>() {
            @Override
            public void onComplete(Task<Location> task) {
                if (task.isSuccessful() && task.getResult() != null) {
                    Location location = task.getResult();
                    // Location found. Create a request:

                    try {
                        SharedPreferences prefs = PreferenceManager.getDefaultSharedPreferences(ctx);
                        boolean mexAllowed = prefs.getBoolean(getResources().getString(R.string.preference_mex_location_verification), false);

                        AppClient.Match_Engine_Request req = mMatchingEngine.createRequest(
                                ctx,
                                location);
                        if (req != null) {
                            // Location Verification (Blocking, or use verifyLocationFuture):
                            AppClient.Match_Engine_Loc_Verify verifiedLocation = mMatchingEngine.verifyLocation(req, 10000);
                            someText = "[Location Verified: Tower: " + verifiedLocation.getTowerStatus() +
                                    ", GPS LocationStatus: " + verifiedLocation.getGpsLocationStatus() + "]\n";

                            // Find the closest cloudlet for your application to use.
                            // (Blocking call, or use findCloudletFuture):
                            FindCloudletResponse closestCloudlet = mMatchingEngine.findCloudlet(req, 10000);
                            // FIXME: It's not possible to get a complete http(s) URI on just a service IP + port!
                            String serverip = null;
                            if (closestCloudlet.server != null && closestCloudlet.server.length > 0) {
                                serverip = closestCloudlet.server[0] + ", ";
                                for (int i = 1; i < closestCloudlet.server.length - 1; i++) {
                                    serverip += closestCloudlet.server[i] + ", ";
                                }
                                serverip += closestCloudlet.server[closestCloudlet.server.length - 1];
                            }
                            someText += "[Cloudlet Server: [" + serverip + "], Port: " + closestCloudlet.port + "]";

                            TextView tv = findViewById(R.id.mex_verified_location_content);
                            tv.setText(someText);
                        } else {
                            someText = "Cannot create request object.";
                            if (!mexAllowed) {
                                someText += " Reason: Enhanced location is disabled.";
                            }
                            TextView tv = findViewById(R.id.mex_verified_location_content);
                            tv.setText(someText);
                        }
                    } catch (StatusRuntimeException sre) {
                        sre.printStackTrace();
                    } catch (IllegalArgumentException iae) {
                        iae.printStackTrace();
                    }
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
}
