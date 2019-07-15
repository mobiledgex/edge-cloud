package com.mobiledgex.matchingengine;

import java.util.Iterator;

import io.grpc.ManagedChannel;

/**
 * Simple Iterator wrapper that keeps a GRPC channel reference alive to read data from that channel.
 * This holds a channel resource until no longer referenced.
 * @param <T>
 */
public class ChannelIterator<T> implements Iterator<T> {

    private ManagedChannel mManagedChannel;
    private Iterator<T> mIterator;

    public ChannelIterator (ManagedChannel channel, Iterator<T> iterator) {
        mManagedChannel = channel;
        mIterator = iterator;
    }

    @Override
    public boolean hasNext() {
        return mIterator.hasNext();
    }

    @Override
    public T next() {
        return mIterator.next();
    }

    @Override
    public void remove() {
        mIterator.remove();
    }
}
