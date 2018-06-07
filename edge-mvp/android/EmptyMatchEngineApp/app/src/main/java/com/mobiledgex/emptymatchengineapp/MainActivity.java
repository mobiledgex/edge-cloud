package com.mobiledgex.emptymatchengineapp;

import android.app.Activity;
import android.location.Location;
import android.os.Bundle;
import android.support.annotation.NonNull;
import android.support.design.widget.FloatingActionButton;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.Toolbar;
import android.util.Log;
import android.view.View;
import android.view.Menu;
import android.view.MenuItem;
import android.widget.TextView;

// Matching Engine API:
import com.mobiledgex.matchingengine.FindCloudletResponse;
import com.mobiledgex.matchingengine.MatchingEngine;
import com.mobiledgex.matchingengine.util.RequestPermissions;

import java.util.concurrent.ExecutionException;

import distributed_match_engine.AppClient;


// Location API:
import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.android.gms.location.LocationServices;
import com.google.android.gms.location.LocationCallback;
import com.google.android.gms.location.LocationRequest;
import com.google.android.gms.location.LocationResult;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;


public class MainActivity extends AppCompatActivity {
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

    /**
     * See documentation for Google's FusedLocationProviderClient for additional usage information.
     */
    private void startLocationUpdates() {
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
        // Or replace with an app specific dialog set.
        mRpUtil.onRequestPermissionsResult(this, requestCode, permissions, grantResults);
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
                    try {

                        // Location found. Create a request:
                        AppClient.Match_Engine_Request req = mMatchingEngine.createRequest(
                                ctx,
                                location);

                        // Location Verification (Blocking, or use verifyLocationFuture):
                        boolean verifiedLocation = mMatchingEngine.verifyLocation(req, 10000);
                        someText = "[Location Verified: " + verifiedLocation + "]\n";


                        // Find the closest cloudlet for your application to use.
                        // (Blocking call, or use findCloudletFuture):
                        FindCloudletResponse closestCloudlet = mMatchingEngine.findCloudlet(req, 10000);
                        // FIXME: It's not possible to get a complete http(s) URI on just a service IP + port!
                        String serverip = null;
                        if (closestCloudlet.server != null && closestCloudlet.server.length > 0) {
                            serverip = closestCloudlet.server[0] + ", ";
                            for (int i = 1; i < closestCloudlet.server.length-1; i++) {
                                serverip += closestCloudlet.server[i] + ", ";
                            }
                            serverip += closestCloudlet.server[closestCloudlet.server.length-1];
                        }
                        someText += "[Cloudlet Server: [" + serverip + "], Port: " + closestCloudlet.port + "]";
                        TextView tv = findViewById(R.id.mex_verified_location_content);
                        tv.setText(someText);
                    } catch (InterruptedException ie) {
                        ie.printStackTrace();
                    } catch (ExecutionException ee) {
                        ee.printStackTrace();
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
