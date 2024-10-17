# Example policy that allows access to the Raft snapshot endpoint
# and reading AWS credentials from a secret
path "sys/storage/raft/snapshot"
{
  capabilities = ["read"]
}

path "<path to secret>"
{
  capabilities = ["read"]
}
