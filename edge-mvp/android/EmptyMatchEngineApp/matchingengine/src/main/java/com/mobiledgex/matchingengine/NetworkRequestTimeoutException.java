package com.mobiledgex.matchingengine;

import java.lang.Exception;

public class NetworkRequestTimeoutException extends Exception {
    public NetworkRequestTimeoutException(String msg) {
        super(msg);
    }
}
