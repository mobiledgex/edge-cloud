package cloudcommon

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
)

func TestDecodeK8S(t *testing.T) {
	objs, _, err := DecodeK8SYaml(testKubeManifest)
	require.Nil(t, err)
	require.Equal(t, 3, len(objs))
	_, ok := objs[0].(*v1.Service)
	require.True(t, ok)
	_, ok = objs[1].(*v1.Service)
	require.True(t, ok)
	_, ok = objs[2].(*appsv1.Deployment)
	require.True(t, ok)

}

var testKubeManifest = `apiVersion: v1
kind: Service
metadata:
  name: testapp-tcp-service
  labels:
    run: testapp
spec:
  type: LoadBalancer
  ports:
  - port: 27272
    targetPort: 27272
    protocol: TCP
    name: grpc27272
  - port: 27273
    targetPort: 27273
    protocol: TCP
    name: rest27273
  - port: 27274
    targetPort: 27274
    protocol: TCP
    name: http27274
  - port: 27275
    targetPort: 27275
    protocol: TCP
    name: tcp27275
  selector:
    run: testapp
---
apiVersion: v1
kind: Service
metadata:
  name: testapp-udp-service
  labels:
    run: testapp
spec:
  type: LoadBalancer
  ports:
  - port: 27276
    targetPort: 27276
    protocol: UDP
    name: udp27276
  selector:
    run: testapp
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testapp-deployment
spec:
  selector:
    matchLabels:
      run: testapp
  replicas: 2
  template:
    metadata:
      labels:
        run: testapp
    spec:
      volumes:
      imagePullSecrets: 
      - name: mexregistrysecret
      containers:
      - name: testapp
        image: registry.mobiledgex.net:5000/mobiledgex/mexexample
        imagePullPolicy: Always
        ports:
        - containerPort: 27272
          protocol: TCP
        - containerPort: 27273
          protocol: TCP
        - containerPort: 27274
          protocol: TCP
        - containerPort: 27275
          protocol: TCP
        - containerPort: 27276
          protocol: UDP
`
