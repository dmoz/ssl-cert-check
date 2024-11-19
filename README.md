

# TLS Certificate Monitor
==========================

A simple Go program that monitors the TLS certificates of a list of websites and sends an email notification when a certificate is near expiration or has expired.

## Features

* Monitors TLS certificates of multiple websites
* Sends email notifications when a certificate is near expiration or has expired
* Uses a configurable threshold for expiration warnings
* Supports custom SMTP server settings

## Configuration

The program uses a JSON configuration file (`config.json`) to store the list of websites to monitor, email settings, and SMTP server settings.

### Example Configuration File

```json
{
  "sites": [
    "example.com",
    "sub.example.com",
    "example.net"
  ],
  "emails": [
    "admin@example.com",
    "dev@example.com"
  ],
  "smtp": {
    "server": "smtp.example.com",
    "from": "tls-monitor@example.com",
    "password": "your-smtp-password"
  },
  "expirationWarningThreshold": "168h",
  "timeout": "5s"
}
```

## Usage

1. Build the program using `go build ssl-cert-check.go`
2. Create a `config.json` file with your desired configuration
3. Run the program using `./ssl-cert-check`

## Notes

* The program uses the default port 443 for HTTPS connections. If a website uses a different port, you can specify it in the `sites` list (e.g. "example.com:8443").
* The program uses a simple email notification system. You can modify the `sendEmail` function to use a different email library or service.
* The program uses a configurable threshold for expiration warnings. You can adjust this value to suit your needs.
* The program uses a configurable timeout for connecting and getting certificates. You can adjust this value to suit your needs.



To build the Docker image, run the following command:

```bash
docker build -t tls-monitor .
```

To run the Docker container, run the following command:

```bash
docker run -d --name tls-monitor tls-monitor
```

Note: You'll need to create a `config.json` file in the same directory as the Dockerfile, or mount a volume with the configuration file when running the container.