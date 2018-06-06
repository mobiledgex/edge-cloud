package com.mobiledgex.matchingengine;

import java.lang.IllegalArgumentException;

public class MissingRequestException extends IllegalArgumentException {
    public MissingRequestException(String msg) {
        super(msg);
    }
}
