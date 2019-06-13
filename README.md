## TFTP Server with Google Cloud Storage backend

Runs an isolated, sandboxed TFTP server that only interacts with virtual backing storage on Google Cloud Storage (GCS).

Also runs a HTTP server from the same bucket, for faster netboots.

Set the following environment variables, e.g.

```
TFTP_ENABLE_HTTP=true
GCS_CREDENTIALS_FILE=/credentials.json
GCS_BUCKET=my-tftp-bucket
```

Read-only operations are supported.