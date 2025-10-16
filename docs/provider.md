# Provider Configuration

The DSPC provider supports configuration through provider blocks and environment variables.

## Configuration Options

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `endpoint` | string | `"http://localhost:8080"` | The endpoint URL for the DSPC VM Deployer API |
| `timeout` | number | `30` | The timeout in seconds for API requests |
| `api_key` | string | `null` | API key for authentication with DSPC API |

## Example Configuration

```hcl
provider "dspc" {
  endpoint = "https://vm-deployer.example.com:8080"
  timeout  = 60
  api_key  = "your-api-key-here"
}
```

## Environment Variables

You can also configure the provider using environment variables:

```bash
export DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
export DSPC_TIMEOUT="60"
export DSPC_API_KEY="your-api-key-here"
```
