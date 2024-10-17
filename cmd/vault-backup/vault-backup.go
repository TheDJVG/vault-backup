package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

const (
	authModeToken      = "token"
	authModeKubernetes = "kubernetes"
)

var (
	authMode                = flag.String("authMode", "token", "Vault authentication mode: token or kubernetes")
	serviceAccountTokenPath = flag.String("kubernetesServiceAccountPath", "/var/run/secrets/kubernetes.io/serviceaccount", "Path to kubernetes service account token")
	vaultMount              = flag.String("mount", "secret", "Vault secret mount")
	secretPath              = flag.String("secret", "", "Path to secret that contains S3 credentials")
)

func main() {

	flag.Parse()

	if *authMode == "" {
		log.Fatal("-authMode not set, set to token or kubernetes")
	}

	vaultConfig := vault.DefaultConfig()
	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		log.Fatal("unable to initialize Vault client: ", err)
	}

	switch *authMode {
	case authModeToken:
		token := os.Getenv("VAULT_TOKEN")

		if token == "" {
			log.Fatal("Vault: env. variable VAULT_TOKEN not set.")
		}
	case authModeKubernetes:
		role := os.Getenv("VAULT_ROLE")

		if role == "" {
			log.Fatal("Vault: env. variable VAULT_ROLE not set.")
		}

		auth, err := auth.NewKubernetesAuth(role, auth.WithServiceAccountTokenPath(*serviceAccountTokenPath))

		if err != nil {
			log.Fatal("Vault: unable to initialize Kubernetes auth method: ", err)
		}

		authInfo, err := client.Auth().Login(context.TODO(), auth)

		if err != nil {
			log.Fatal(err)
		}
		if authInfo == nil {
			log.Fatal("Vault: no auth info was returned after login")

		}

	default:
		log.Fatalf("ERROR: authMode '%s' unknown, set to token or kubernetes", *authMode)
	}

	if *secretPath != "" {
		secret, err := client.KVv2(*vaultMount).Get(context.Background(), *secretPath)
		if err != nil {
			log.Fatal("unable to read Vault secret:", err)
		}

		// Set all the keys to env. variables ussed by the AWS SDK
		for k, v := range secret.Data {
			switch value := v.(type) {
			case string:
				log.Printf("%s env. variable set from secret", k)
				os.Setenv(k, value)
			default:
				log.Printf("Warning: cannot set env. '%s' as type '%s' is not a string", k, reflect.TypeOf(v))
			}
		}
	}

	bucketName := os.Getenv("AWS_BUCKET")

	if bucketName == "" {
		log.Fatal("'AWS_BUCKET' not set")
	}

	// // Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Unable to init AWS SDK:", err)
	}

	pathStyle, err := strconv.ParseBool(os.Getenv("AWS_PATHSTYLE"))
	if err != nil {
		pathStyle = false
	}
	endpoint := os.Getenv("AWS_ENDPOINT")

	// // Create an S3 client
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
		o.UsePathStyle = pathStyle
	})

	// Pipe straight from Vault to S3
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		err = client.Sys().RaftSnapshotWithContext(context.TODO(), writer)

		if err != nil {
			log.Fatal("unable to read snapshot from Vault: ", err)
		}
		log.Println("Vault snapshot created")

	}()

	filename := time.Now().Format("2006_01_02__15_04.raft")

	// Upload file to S3 using the manager as the lenght might not be known yet.
	uploader := manager.NewUploader(s3Client)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   reader,
	})
	if err != nil {
		log.Fatal("failed to upload file: ", err)
	}

	log.Printf("Vault snapshot uploaded as %s", filename)
}
