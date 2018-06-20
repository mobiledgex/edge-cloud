package com.mobiledgec.markcloudlets;

import android.Manifest;
import android.content.Context;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.location.Address;
import android.location.Geocoder;
import android.location.Location;
import android.location.LocationListener;
import android.location.LocationManager;
import android.os.AsyncTask;
import android.os.Build;
import android.os.Bundle;
import android.support.annotation.NonNull;
import android.support.v4.app.ActivityCompat;
import android.support.v4.app.FragmentActivity;
import android.support.v4.content.ContextCompat;
import android.util.Log;
import android.widget.Toast;

import com.google.android.gms.maps.CameraUpdateFactory;
import com.google.android.gms.maps.GoogleMap;
import com.google.android.gms.maps.OnMapReadyCallback;
import com.google.android.gms.maps.SupportMapFragment;
import com.google.android.gms.maps.model.BitmapDescriptorFactory;
import com.google.android.gms.maps.model.LatLng;
import com.google.android.gms.maps.model.MarkerOptions;
import com.google.android.gms.maps.model.PolylineOptions;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

public class MapsActivity extends FragmentActivity implements OnMapReadyCallback {

    private GoogleMap mMap;
    List<PointsOfInterest> mPOIResult = new ArrayList<>();

    LocationManager locationManager;
    LocationListener locationListener;

    @Override
    public void onRequestPermissionsResult(int requestCode, @NonNull String[] permissions, @NonNull int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);

        if (requestCode == 1) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED) {

                    locationManager.requestLocationUpdates(LocationManager.GPS_PROVIDER, 0, 0, locationListener);
                }
            }
        }
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_maps);
        // Obtain the SupportMapFragment and get notified when the map is ready to be used.
        SupportMapFragment mapFragment = (SupportMapFragment) getSupportFragmentManager()
                .findFragmentById(R.id.map);
        mapFragment.getMapAsync(this);


        DownloadTask task = new DownloadTask();
        //task.execute("https://api.myjson.com/bins/6u1n2");
        task.execute("https://api.myjson.com/bins/11dtje");

    }

    public class DownloadTask extends AsyncTask<String, Void, String> {

        @Override
        protected String doInBackground(String... urls) {

            String result = "";
            URL url;
            HttpURLConnection urlConnection = null;

            try {
                url = new URL(urls[0]);

                urlConnection = (HttpURLConnection) url.openConnection();

                InputStream in = urlConnection.getInputStream();

                InputStreamReader reader = new InputStreamReader(in);

                int data = reader.read();

                while (data != -1) {

                    char current = (char) data;

                    result += current;

                    data = reader.read();

                }

                return result;

            } catch (MalformedURLException e) {
                e.printStackTrace();
            } catch (IOException e) {
                e.printStackTrace();
            }

            return null;
        }

        @Override
        protected void onPostExecute(String result) {
            super.onPostExecute(result);

            try {

                JSONObject cloudletObject = new JSONObject(result);
                String cloudletsInfo = cloudletObject.getString("cloudlets");
                JSONArray arr = new JSONArray(cloudletsInfo);

                String[] latitude = new String[arr.length()];
                String[] longitude = new String[arr.length()];
                String[] cloudlets = new String[arr.length()];
                String[] operators = new String[arr.length()];


                mPOIResult = new ArrayList<>();
                System.out.print(arr.length());
                for (int i = 0; i < arr.length(); i++) {

                    JSONObject jsonPart = arr.getJSONObject(i);
                    latitude[i] = jsonPart.getString("lat");
                    longitude[i] = jsonPart.getString("lng");
                    cloudlets[i] = jsonPart.getString("cloudlet");
                    operators[i] = jsonPart.getString("operator");

                    PointsOfInterest poi = new PointsOfInterest(
                            Double.parseDouble(latitude[i]),
                            Double.parseDouble(longitude[i]),
                            cloudlets[i],
                            operators[i]);

                   mPOIResult.add(poi);
                }


            } catch (JSONException e) {
                e.printStackTrace();
            }

        }

    }



    /**
     * Manipulates the map once available.
     * This callback is triggered when the map is ready to be used.
     * This is where we can add markers or lines, add listeners or move the camera.
     * If Google Play services is not installed on the device, the user will be prompted to install
     * it inside the SupportMapFragment. This method will only be triggered once the user has
     * installed Google Play services and returned to the app.
     */
    @Override
    public void onMapReady(final GoogleMap googleMap) {
        mMap = googleMap;

        locationManager = (LocationManager) this.getSystemService(Context.LOCATION_SERVICE);
        locationListener = new LocationListener() {
            @Override
            public void onLocationChanged(Location location) {

                //Toast.makeText(MapsActivity.this, location.toString(), Toast.LENGTH_SHORT).show();
                mMap.setMapType(GoogleMap.MAP_TYPE_HYBRID);
                LatLng userLocation = new LatLng(location.getLatitude(),location.getLongitude());
                mMap.clear();

                LatLng menlopark = new LatLng(37.4530, -122.1817);
                LatLng paloalto = new LatLng(37.4419,-122.1430 );
                LatLng mountainview = new LatLng (37.3861, -122.0839);

                float [] distMP = new float[1];
                float [] distPA = new float[1];
                float [] distMV = new float[1];
                float minimumDistance = Float.MAX_VALUE;
                float[] currentDistance = new float[1];
                String closestCloudlet = "";
                ArrayList<LatLng> polyLine = new ArrayList<>();


                mMap.addMarker(new MarkerOptions().position(userLocation).title("Your Location").icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_RED)));

                for (PointsOfInterest poi : mPOIResult) {

                    double lat = poi.getLatitude();
                    Log.i("DDDDDDD", poi.getOperators() + poi.getCloudlets());
                    double lng = poi.getLongitude();
                    LatLng l = new LatLng(lat, lng);

                    Location.distanceBetween(userLocation.latitude, userLocation.longitude,lat,lng, currentDistance);
                    if (minimumDistance > currentDistance[0]) {
                        minimumDistance = currentDistance[0];
                        closestCloudlet =  poi.getCloudlets();


                    }

                    if (poi.getOperators().equals("TMOBILE")) {

                        mMap.addMarker(new MarkerOptions()
                                .position(l)
                                .title(poi.getCloudlets())
                                .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_BLUE)));

                    } else if (poi.getOperators().equals("ATT")) {

                        mMap.addMarker(new MarkerOptions()
                                .position(l)
                                .title(poi.getCloudlets())
                                .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_GREEN)));

                    } else if (poi.getOperators().equals("VERIZON")) {

                        mMap.addMarker(new MarkerOptions()
                                .position(l)
                                .title(poi.getCloudlets())
                                .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_CYAN)));

                    }

                    polyLine.add(l);
                    polyLine.add(userLocation);

                    PolylineOptions poly = new PolylineOptions().width(5).color(Color.WHITE).geodesic(true);
                    for (int z = 0; z < polyLine.size(); z++) {
                        LatLng point = polyLine.get(z);
                        poly.add(point);
                    }
                    mMap.addPolyline(poly);
                }

                //Toast.makeText(MapsActivity.this,  address, Toast.LENGTH_LONG).show();
                Toast.makeText(MapsActivity.this, "Closest Cloudlet from Userlocation = " + String.valueOf(minimumDistance) + "meters", Toast.LENGTH_LONG).show();
                Toast.makeText(MapsActivity.this, "Closest Cloudlet from Userlocation = " + closestCloudlet, Toast.LENGTH_LONG).show();



                Geocoder geocoder = new Geocoder(getApplicationContext(), Locale.getDefault()); //LOCALE IS FORMAT OF ADDRESS FOR DIF COUNTRIES

                try {

                    List<Address> listAddress = geocoder.getFromLocation(location.getLatitude(), location.getLongitude(), 1);

                    if (listAddress != null && listAddress.size() > 0) {

                        Log.i("PlaceInfo", listAddress.get(0).toString());

                        String address = "";
                        if (listAddress.get(0).getSubThoroughfare() != null) {
                            address += listAddress.get(0).getSubThoroughfare() + ", ";
                        }

                        if (listAddress.get(0).getThoroughfare() != null) {
                            address += listAddress.get(0).getThoroughfare() + ", ";
                        }

                        if (listAddress.get(0).getThoroughfare() != null) {
                            address += listAddress.get(0).getThoroughfare() + ", ";
                        }

                        if (listAddress.get(0).getPostalCode() != null) {
                            address += listAddress.get(0).getPostalCode() + " ";
                        }

                        if (listAddress.get(0).getCountryName() != null) {
                            address += listAddress.get(0).getCountryName();
                        }

                        //Toast.makeText(MapsActivity.this,  address, Toast.LENGTH_LONG).show();

                    }
                } catch (IOException e) {

                    e.printStackTrace();
                }

                mMap.moveCamera(CameraUpdateFactory.newLatLngZoom(userLocation, 11));
            }

            @Override
            public void onStatusChanged(String s, int i, Bundle bundle) {

            }

            @Override
            public void onProviderEnabled(String s) {

            }

            @Override
            public void onProviderDisabled(String s) {

            }
        };

        if (Build.VERSION.SDK_INT < 23) {
            locationManager.requestLocationUpdates(LocationManager.GPS_PROVIDER,0,0, locationListener);
        } else {

            if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED) {
                ActivityCompat.requestPermissions(this, new String[]{Manifest.permission.ACCESS_FINE_LOCATION}, 1);

            } else {
                locationManager.requestLocationUpdates(LocationManager.GPS_PROVIDER,0,0, locationListener);

                Location lastKnownLocation = locationManager.getLastKnownLocation(LocationManager.GPS_PROVIDER);

                LatLng userLocation = new LatLng(lastKnownLocation.getLatitude(), lastKnownLocation.getLongitude());
                LatLng menlopark = new LatLng(37.4530, -122.1817);
                LatLng paloalto = new LatLng(37.4419,-122.1430 );
                LatLng mountainview = new LatLng (37.3861, -122.0839);


                mMap.clear();
                mMap.addMarker(new MarkerOptions().position(userLocation).title("Your Location").icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_RED)));
                mMap.moveCamera(CameraUpdateFactory.newLatLngZoom(userLocation, 11));
            }
        }
    }
}
