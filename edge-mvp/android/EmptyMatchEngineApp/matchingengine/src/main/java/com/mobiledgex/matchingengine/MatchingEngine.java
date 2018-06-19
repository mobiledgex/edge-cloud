package com.mobiledgex.matchingengine;

import android.content.Context;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.telephony.NeighboringCellInfo;
import android.telephony.TelephonyManager;

import com.google.protobuf.ByteString;

import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.AppClient.Match_Engine_Request;
import distributed_match_engine.LocOuterClass.Loc;
import io.grpc.StatusRuntimeException;

import android.content.pm.PackageInfo;


// TODO: GRPC (which needs http/2).
public class MatchingEngine {
    public static final String TAG = "MatchingEngine";
    private String host = "192.168.28.162"; // FIXME: Your available external server IP until the real server is up.
    //private String host = "192.168.1.91"; // FIXME: Your available external server IP until the real server is up.
    private int port = 50051;

    // A threadpool for all the MatchEngine API callable interfaces:
    final ExecutorService threadpool;

    public MatchingEngine() {
        threadpool = Executors.newSingleThreadExecutor();
    }
    public MatchingEngine(ExecutorService executorService) {
        threadpool = executorService;
    }

    public Future submit(Callable task) {
        return threadpool.submit(task);
    }

    /**
     * The library itself will not directly ask for permissions, the application should before use.
     * This keeps responsibilities managed clearly in one spot under the app's control.
     */
    public Match_Engine_Request createRequest(Context context, android.location.Location loc) throws SecurityException {
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        // Operator
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String telName = telManager.getNetworkOperatorName();
        // READ_PHONE_STATE or
        Match_Engine_Request.IDType id_type = Match_Engine_Request.IDType.MSISDN;
        String id = telManager.getLine1Number(); // NOT IMEI, if this throws a SecurityException, application must handle it.
        String mnc = telManager.getNetworkOperator();
        String mcc = telManager.getNetworkCountryIso();

        if (id == null) { // Fallback to IP:
	    // TODO: Dual SIM?
        }

        // Tower
        TelephonyManager tm = (TelephonyManager) context.getSystemService(Context.TELEPHONY_SERVICE);
        List<NeighboringCellInfo> neighbors = tm.getNeighboringCellInfo();
        int lac = 0;
        int cid = 0;
        if (neighbors.size() > 0) {
            lac = neighbors.get(0).getLac();
            cid = neighbors.get(0).getCid();
        }

        // App
        ApplicationInfo appInfo = context.getApplicationInfo();
        String packageLabel = "";
        if (context.getPackageManager() != null) {
            CharSequence seq = appInfo.loadLabel(context.getPackageManager());
            if (seq != null) {
                packageLabel = seq.toString();
            }
        }
        PackageInfo pInfo;
        String versionName = "";
        String versionCode = "";
        String appName = packageLabel;
        try {
            pInfo = context.getPackageManager().getPackageInfo(context.getPackageName(), 0);
            versionName = pInfo.versionName;
            versionCode = (new Integer(pInfo.versionCode)).toString();
        } catch (PackageManager.NameNotFoundException nfe) {
            nfe.printStackTrace();
            // Hard stop, or continue?
        }

        // Passed in Location (which is a callback interface)
        Loc aLoc = Loc.newBuilder()
                .setLat((loc == null) ? 0.0d : loc.getLatitude())
                .setLong((loc == null) ? 0.0d : loc.getLongitude())
                .setHorizontalAccuracy((loc == null) ? 0.0d :loc.getAccuracy())
                //.setVerticalAccuracy(loc.getVerticalAccuracyMeters()) // API Level 26 required.
                .setVerticalAccuracy(0d)
                .setAltitude((loc == null) ? 0.0d : loc.getAltitude())
                .setCourse((loc == null) ? 0.0d : loc.getBearing())
                .setSpeed((loc == null) ? 0.0d : loc.getSpeed())
                .build();

        Match_Engine_Request request = AppClient.Match_Engine_Request.newBuilder()
                .setVer(5)
                .setIdType(id_type)
                .setId((id == null) ? "" : id)
                .setCarrierID(3l) // uint64 --> String? mnc, mcc?
                .setCarrierName(mnc.equals("") ? "" : mnc) // Mobile Network Carrier
                .setTower(cid) // cid and lac (int)
                .setGpsLocation(aLoc)
                .setAppId(5011l) // uint64 --> String again. TODO: Clarify use.
                .setProtocol(ByteString.copyFromUtf8("http")) // This one is appId context sensitive.
                .setServerPort(ByteString.copyFromUtf8("1234")) // App dependent.
                .setDevName(packageLabel) // From signing certificate?
                .setAppName(appName)
                .setAppVers(versionCode) // Or versionName, which is visual name?
                .setToken("") // None.
                .build();


        return request;
    }

    /**
     * Registers Client using blocking API call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     */
    public AppClient.Match_Engine_Status registerClient(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
        RegisterClient registerClient = new RegisterClient(this);
        registerClient.setRequest(request, timeoutInMilliseconds);
        return registerClient.call();
    }

    /**
     * Registers device on the MatchingEngine server. Returns a Future.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppClient.Match_Engine_Status> registerClientFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        RegisterClient registerClient = new RegisterClient(this);
        registerClient.setRequest(request, timeoutInMilliseconds);
        return submit(registerClient);
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request.
     * @param request
     * @return cloudlet URI.
     * @throws StatusRuntimeException
     */
    public FindCloudletResponse findCloudlet(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, timeoutInMilliseconds);
        return findCloudlet.call();
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request. Returns a Future.
     * @param request
     * @return cloudlet URI Future.
     */
    public Future<FindCloudletResponse> findCloudletFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, timeoutInMilliseconds);
        return submit(findCloudlet);
    }


    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side.
     * @param request
     * @return boolean validated or not.
     * @throws StatusRuntimeException
     */
    public AppClient.Match_Engine_Loc_Verify verifyLocation(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, timeoutInMilliseconds);
        return verifyLocation.call();
    }

    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side. Returns a future.
     * @param request
     * @return Future<Boolean> validated or not.
     */
    public Future<AppClient.Match_Engine_Loc_Verify> verifyLocationFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, timeoutInMilliseconds);
        return submit(verifyLocation);
    }

    /**
     * getLocation returns the network verified location of this device.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     */
    public AppClient.Match_Engine_Loc getLocation(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, timeoutInMilliseconds);
        return getLocation.call();
    }

    /**
     * getLocation returns the network verified location of this device. Returns a Future.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppClient.Match_Engine_Loc> getLocationFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, timeoutInMilliseconds);
        return submit(getLocation);
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }
}
