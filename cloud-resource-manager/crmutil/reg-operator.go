package crmutil

import (
	"context"
	"log"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func RegisterOperators(api edgeproto.OperatorApiClient) error {
	var err error

	ctx := context.TODO()

	for _, op := range OperatorData {
		_, err = api.CreateOperator(ctx, &op)
		if err != nil {
			log.Printf("createOperator failed %v", err)
			return err
		}
	}

	return nil
}
