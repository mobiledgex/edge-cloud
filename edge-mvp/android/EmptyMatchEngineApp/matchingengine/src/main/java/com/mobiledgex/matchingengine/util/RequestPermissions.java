package com.mobiledgex.matchingengine.util;

import android.Manifest;
import android.app.Activity;
import android.app.AlertDialog;
import android.app.Dialog;
import android.content.DialogInterface;
import android.content.pm.PackageManager;
import android.os.Bundle;
import android.support.annotation.NonNull;
import android.support.v4.app.DialogFragment;
import android.support.v4.app.Fragment;
import android.support.v7.app.AppCompatActivity;
import android.support.v4.app.ActivityCompat;
import android.support.v4.content.ContextCompat;

import java.util.ArrayList;
import java.util.List;

import com.mobiledgex.matchingengine.R;

/**
 * Android UI Permissions helper. Activity contexts are needed.
 */
public class RequestPermissions {
    public static final int REQUEST_MULTIPLE_PERMISSION = 1001;
    public static final String[] permissions = new String[] { // Special Enhanced security requests.
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
            Manifest.permission.READ_PHONE_STATE, // Get Phone number
    };

    public static boolean permissionsDeniedShown = false;

    public List<String> getNeededPermissions(Activity activity) {
        List<String> permissionsNeeded = new ArrayList<>();

        for (String pStr : permissions) {
            int result = ContextCompat.checkSelfPermission(activity, pStr);
            if (result != PackageManager.PERMISSION_GRANTED) {
                permissionsNeeded.add(pStr);
            }
        }
        return permissionsNeeded;
    }

    public void requestMultiplePermissions(AppCompatActivity activity) {
        // Check which ones missing
        List<String> permissionsNeeded = getNeededPermissions(activity);

        String[] permissionArray;
        if (!permissionsNeeded.isEmpty()) {
            permissionArray = permissionsNeeded.toArray(new String[permissionsNeeded.size()]);
        } else {
            permissionArray = permissions; // Nothing was granted. Ask for all.
        }

        ActivityCompat.requestPermissions(activity, permissionArray, REQUEST_MULTIPLE_PERMISSION);
    }

    /**
     * Keeps asking for permissions until granted or user checks box to not asked again.
     * @param activity
     * @param requestCode
     * @param permissions
     * @param grantResults
     */
    public void onRequestPermissionsResult(AppCompatActivity activity, int requestCode, @NonNull String[] permissions,
                                           @NonNull int[] grantResults) {

        int numGranted = 0;
        boolean showWarning = false;
        for (int i = 0; i < grantResults.length; i++) {
            if (grantResults[i] == PackageManager.PERMISSION_GRANTED) {
                numGranted++;
            }
        }
        showWarning = (numGranted != grantResults.length);

        if (requestCode == REQUEST_MULTIPLE_PERMISSION) {
            for (int i = 0; i < grantResults.length; i++) {
                if (ActivityCompat.shouldShowRequestPermissionRationale(activity, permissions[i])) {
                    new ConfirmationDialog().show(activity.getSupportFragmentManager(), "dialog");
                } else if (!permissionsDeniedShown && showWarning) {
                    // Rejected, or user asks to not ask again. This may still be critical for the
                    // application, so show once.
                    String msg = activity.getResources().getString(R.string.request_permission);
                    ErrorDialog.newInstance(msg).show(activity.getSupportFragmentManager(), "errorDialog");
                }
                return;
            }
        }
    }


    /**
     * Shows OK/Cancel confirmation dialog about needed permissions.
     */
    public static class ConfirmationDialog extends DialogFragment {

        @NonNull
        @Override
        public Dialog onCreateDialog(Bundle savedInstanceState) {
            final Fragment parent = this;
            return new AlertDialog.Builder(getActivity())
                    .setMessage(R.string.request_permission)
                    .setPositiveButton(android.R.string.ok, new DialogInterface.OnClickListener() {
                        @Override
                        public void onClick(DialogInterface dialog, int which) {
                            requestPermissions(permissions, REQUEST_MULTIPLE_PERMISSION);
                            dialog.dismiss();
                        }
                    })
                    .setNegativeButton(android.R.string.cancel,
                            new DialogInterface.OnClickListener() {
                                @Override
                                public void onClick(DialogInterface dialog, int which) {
                                    dialog.dismiss();
                                }
                            })
                    .create();
        }
    }

    /**
     * Shows an error message dialog.
     */
    public static class ErrorDialog extends DialogFragment {

        private static final String ARG_MESSAGE = "message";

        public static ErrorDialog newInstance(String message) {
            ErrorDialog dialog = new ErrorDialog();
            Bundle args = new Bundle();
            args.putString(ARG_MESSAGE, message);
            dialog.setArguments(args);
            return dialog;
        }

        @NonNull
        @Override
        public Dialog onCreateDialog(Bundle savedInstanceState) {
            final Activity activity = getActivity();
            return new AlertDialog.Builder(activity)
                    .setMessage(getArguments().getString(ARG_MESSAGE))
                    .setPositiveButton(android.R.string.ok, new DialogInterface.OnClickListener() {
                        @Override
                        public void onClick(DialogInterface dialogInterface, int i) {
                            permissionsDeniedShown = true;
                            dialogInterface.dismiss();
                        }
                    })
                    .create();
        }

    }
}
