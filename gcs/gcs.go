package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/mobiledgex/edge-cloud/vault"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCSClient struct {
	Client     *storage.Client
	BucketName string
	Timeout    time.Duration
	AuthConf   *jwt.Config
}

const (
	GCSVaultPath = "secret/data/accounts/gcs"
	ShortTimeout = 5 * time.Minute
	LongTimeout  = 20 * time.Minute

	NotFoundError = "object doesn't exist"
)

func GetGCSCreds(ctx context.Context, vaultConfig *vault.Config) ([]byte, error) {
	client, err := vaultConfig.Login()
	if err != nil {
		return nil, err
	}
	gcsData, err := vault.GetKV(client, GCSVaultPath, 0)
	if err != nil {
		return nil, err
	}
	credsObj, err := json.Marshal(gcsData["data"])
	if err != nil {
		return nil, err
	}
	return credsObj, nil
}

// Must call GCSClient.Close() when done
func NewClient(ctx context.Context, credsObj []byte, bucketName string, timeout time.Duration) (*GCSClient, error) {
	gcsClient := &GCSClient{}

	conf, err := google.JWTConfigFromJSON(credsObj)
	if err != nil {
		return nil, fmt.Errorf("Failed to read service account credentials: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(credsObj))
	if err != nil {
		return nil, fmt.Errorf("Failed to create new GCS client: %v", err)
	}

	gcsClient.Client = storageClient
	gcsClient.BucketName = bucketName
	gcsClient.Timeout = timeout
	gcsClient.AuthConf = conf
	return gcsClient, nil
}

func (gc *GCSClient) Close() {
	if gc != nil && gc.Client != nil {
		gc.Client.Close()
	}
}

// Returns empty if no objects present
func (gc *GCSClient) ListObjects(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	objs := []string{}
	it := gc.Client.Bucket(gc.BucketName).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to list bucket(%q) objects: %v", gc.BucketName, err)
		}
		objs = append(objs, attrs.Name)
	}
	return objs, nil
}

// Overwrites object if already exists
func (gc *GCSClient) UploadObject(ctx context.Context, objectName string, buf *bytes.Buffer) error {
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := gc.Client.Bucket(gc.BucketName).Object(objectName).NewWriter(ctx)
	reader := bytes.NewReader(buf.Bytes())
	if _, err := io.Copy(wc, reader); err != nil {
		return fmt.Errorf("Failed to read object to upload to %s: %v", objectName, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Failed to write object(%q) to GCS: %v", objectName, err)
	}
	return nil
}

func (gc *GCSClient) DownloadObject(ctx context.Context, objectName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	rc, err := gc.Client.Bucket(gc.BucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to download GCS object(%q): %v", objectName, err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("Failed to read download object(%q): %v", objectName, err)
	}
	return data, nil
}

func (gc *GCSClient) DeleteObject(ctx context.Context, objectName string) error {
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	o := gc.Client.Bucket(gc.BucketName).Object(objectName)
	if err := o.Delete(ctx); err != nil {
		if !strings.Contains(err.Error(), NotFoundError) {
			return fmt.Errorf("Failed to delete GCS object(%q): %v", objectName, err)
		}
	}
	return nil
}

// Generates object signed URL with GET method. This gives time-limited resource access.
func (gc *GCSClient) GenerateV4GetObjectSignedURL(ctx context.Context, objectName string, validity time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ShortTimeout)
	defer cancel()
	objs, err := gc.ListObjects(ctx)
	if err != nil {
		return "", err
	}
	found := false
	for _, objName := range objs {
		if objName == objectName {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("Unable to find object(%q) in bucket(%q)", objectName, gc.BucketName)
	}
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		GoogleAccessID: gc.AuthConf.Email,
		PrivateKey:     gc.AuthConf.PrivateKey,
		Expires:        time.Now().Add(validity),
	}
	urlStr, err := storage.SignedURL(gc.BucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("Failed to generate signed URL for bucket(%q), object(%q): %v", gc.BucketName, objectName, err)
	}
	return urlStr, nil
}
