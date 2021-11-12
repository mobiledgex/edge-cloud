package k8smgmt

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"k8s.io/apimachinery/pkg/api/resource"
)

// QuantityToUdec64 converts a kubernetes Quantity to a Udec64.
// This should only fail if the value is negative, or the Quantity
// does not conform to the range of values it's supposed to be limited to.
func QuantityToUdec64(q resource.Quantity) (*edgeproto.Udec64, error) {
	valInt64, ok := q.AsInt64()
	if ok {
		if valInt64 < 0 {
			return nil, fmt.Errorf("Cannot assign negative quantity %s to unsigned decimal", q.String())
		}
		return &edgeproto.Udec64{
			Whole: uint64(valInt64),
		}, nil
	}
	dec := q.AsDec()
	if dec.Sign() < 0 {
		return nil, fmt.Errorf("Cannot assign negative quantity %s to unsigned decimal", q.String())
	}
	scale := int(dec.Scale())
	// Quantity will never have a value larger than 2^63-1, or more than
	// 3 decimal places. So dec.Unscaled() should never fail.
	unscaled, ok := dec.Unscaled()
	if !ok {
		return nil, fmt.Errorf("Unexpected quantity %s out of range", q.String())
	}
	if scale == 0 {
		return &edgeproto.Udec64{
			Whole: uint64(unscaled),
		}, nil
	}
	if scale < 0 {
		// whole number, scale up value
		upscale := uint64(1)
		for i := 0; i > scale; i-- {
			upscale *= 10
		}
		return &edgeproto.Udec64{
			Whole: uint64(unscaled) * upscale,
		}, nil
	}
	// scale > 0, means decimal value
	downscale := uint64(1)
	for i := 0; i < scale; i++ {
		downscale *= 10
	}
	whole := uint64(unscaled) / downscale
	decimals := uint64(unscaled) % downscale
	// decimals is value after decimal place, so
	// 1.2045 would have 2045 as the decimal. This needs to
	// be scaled to nanos.
	for i := edgeproto.DecPrecision; i > scale; i-- {
		decimals *= 10
	}
	return &edgeproto.Udec64{
		Whole: whole,
		Nanos: uint32(decimals),
	}, nil
}

// QuantityToUint64 converts a kubernetes Quantity to a uint64.
// This fails if the quantity is negative or has decimal precision.
func QuantityToUint64(q resource.Quantity) (uint64, error) {
	valInt64, ok := q.AsInt64()
	if !ok {
		return 0, fmt.Errorf("Cannot convert quantity %s to uint64", q.String())
	}
	if valInt64 < 0 {
		return 0, fmt.Errorf("Cannot assign negative quantity %s to uint64", q.String())
	}
	return uint64(valInt64), nil
}
