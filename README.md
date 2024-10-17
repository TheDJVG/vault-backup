# vault-backup
Simple tool to backup a Vault raft snapshot to S3 or S3 compatible endpoint.

## Install
```
go install github.com/thedjvg/vault-backup/cmd/vault-backup@latest
```

Or build the container (currently no public image provided)
```
docker build -t vault-backup .
```

## Configuration

### Vault
Takes the standard `VAULT_*` environment variables as allowed with `VAULT_ADDR` being mandatory.
When `-authMode token` is specicied `VAULT_TOKEN` is required.

When `-authMode` is set to `kubernetes` a `VAULT_ROLE` environment variable should be set. Optionally `-kubernetesServiceAccountPath` can be specified when a service account is set to a different location.

The token needs to have access to raft snapshot endpoint and optionally a secret with S3 credentials. See [examples/policy.hcl](examples/policy.hcl) for an example.

### S3
Takes the standard `AWS_*` environment variables and aditionally accepts the following:

| Name | Description |
|-|-|
| AWS_BUCKET | Bucket to store snapshot |
| AWS_ENDPOINT | Endpoint when not using AWS S3|
| AWS_PATHSTYLE | True when pathstyle buckets are used (e.g. with a different endpoint)

The credentials can be read from Vault using the `-mount` and `-secret` arguments

## Example output
```bash
./vault-backup -authMode token -mount kv -secret apps/aws/vault-backup
2024/10/17 11:10:56 AWS_ACCESS_KEY env. variable set from secret
2024/10/17 11:10:56 AWS_BUCKET env. variable set from secret
2024/10/17 11:10:56 AWS_ENDPOINT env. variable set from secret
2024/10/17 11:10:56 AWS_PATHSTYLE env. variable set from secret
2024/10/17 11:10:56 AWS_REGION env. variable set from secret
2024/10/17 11:10:56 AWS_SECRET_KEY env. variable set from secret
2024/10/17 11:10:56 Vault snapshot created
2024/10/17 11:10:56 Vault snapshot uploaded as 2024_10_17__11_10.raft
```
## Help
```bash
./vault-backup -h
Usage of ./vault-backup:
  -authMode string
        Vault authentication mode: token or kubernetes (default "token")
  -kubernetesServiceAccountPath string
        Path to kubernetes service account token (default "/var/run/secrets/kubernetes.io/serviceaccount")
  -mount string
        Vault secret mount (default "secret")
  -secret string
        Path to secret that contains S3 credentials
````
