package com.mobiledgex.matchingengine.util;

import android.content.Context;
import android.location.Location;

import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.android.gms.location.LocationServices;
import com.google.android.gms.tasks.OnCanceledListener;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;
import com.mobiledgex.matchingengine.MatchingEngine;

import java.lang.ref.WeakReference;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import android.util.Log;


// Simple util class for last known location.
public class MeLocation {
    private static final String TAG = "MeLocation";
    private MatchingEngine mMatchingEngine;
    private Location mLocation;
    private volatile boolean mWaitingForNotify = false;
    private Object syncObject = new Object();
    private long mTimeoutInMilliseconds;

    private class LocationCallable implements Callable<Location> {
        WeakReference<Context> mContext;
        LocationCallable(final Context context) {
            mContext = new WeakReference<>(context);
        }
        public Location call() throws InterruptedException {

            FusedLocationProviderClient fusedLocationClient;
            Context context = mContext.get();
            if (context == null) {
                // Context dead
                return null;
            }

            mWaitingForNotify = true;
            mLocation = null;
            fusedLocationClient = LocationServices.getFusedLocationProviderClient(context);
            try {
                fusedLocationClient.getLastLocation().addOnCompleteListener(new OnCompleteListener<Location>() {
                    @Override
                    public void onComplete(Task<Location> task) {
                        if (task.isSuccessful() && task.getResult() != null) {
                            Location loc = task.getResult();
                            mLocation = loc;
                            synchronized (syncObject) {
                                mWaitingForNotify = false;
                                syncObject.notify();
                            }
                        } else {
                            synchronized (syncObject) {
                                mWaitingForNotify = false;
                                syncObject.notify();
                            }
                            if (task.getException() != null) {
                                Log.w(TAG, "getLastLocation: Exception: ", task.getException());
                            }
                        }
                    }
                });
                fusedLocationClient.getLastLocation().addOnCanceledListener(new OnCanceledListener() {
                    @Override
                    public void onCanceled() {
                        synchronized (syncObject) {
                            mWaitingForNotify = false;
                            syncObject.notify();
                        }
                    }
                });
                synchronized (syncObject) {
                    long timeStart = System.currentTimeMillis();
                    long elapsed = 0;
                    while (mWaitingForNotify == true &&
                            (elapsed = System.currentTimeMillis() - timeStart) < mTimeoutInMilliseconds) {
                        syncObject.wait(mTimeoutInMilliseconds - elapsed);
                    }
                }
            } catch (SecurityException se) {
                mLocation = null;
                throw se;
            } catch (InterruptedException ie) {
                mLocation = null;
                throw ie;
            }

            return mLocation;
        }
    }


    public MeLocation(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    /**
     * A utility blocking call to location services, otherwise, use standard asynchronous location APIs.
     * Location Access Permissions must be enabled.
     */
    public android.location.Location getBlocking(final Context context, long timeoutInMilliseconds)
            throws IllegalArgumentException, IllegalStateException, SecurityException,
                InterruptedException, ExecutionException {
        if (context == null) {
            throw new IllegalStateException("Location util requires a Context.");
        }

        if (timeoutInMilliseconds < 0) {
            throw new IllegalArgumentException("Timeout must be higher than 0.");
        }

        try {
            Callable loc = new LocationCallable(context);
            mTimeoutInMilliseconds = timeoutInMilliseconds;
            Future<Location> locFuture = mMatchingEngine.submit(loc);
            mLocation = locFuture.get();

        } catch (InterruptedException ie) {
            mLocation = null;
            throw ie;
        } catch (ExecutionException ee) {
            mLocation = null;
            throw ee;
        }

        return mLocation;
    }

}
