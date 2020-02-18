# sigsink

A http signature validation server.

# Usage

1. Generate an RSA key and place the public key in the `./keys` directory.

   ```
   $ openssl genrsa -out httpsig.pem 2048
   $ openssl rsa -in httpsig.pem -outform PEM -pubout -out ./keys/httpsig-public.pem
   ```
   
   Keys loaded from files must have a `key_id` header in the PEM block:
   
   ```pem
   -----BEGIN PUBLIC KEY-----
   key_id: httpsig
   
   MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA16tbOpoWq9fhxxFHm/Io
   <key stuff>
   -----END PUBLIC KEY-----

   ```

2. Start the server

   ```
   go run main.go --domain whatever.ngrok.io
   ```

3. Submit a signed POST request to port 7000.

# Docker Usage

1. Build the container

   $ DOCKER_BUILDKIT=1 docker build -t sigsink:1 .

2. Run the container

   $ docker run --mount type=bind,source=/path/to/keys/,target=/sigsink/keys/ -p 7000:7000 -t sigsink:1

# Options

    NAME:
       sigsink - An http signature verification tool.
    
    USAGE:
       sigsink [global options] command [command options] [arguments...]
    
    COMMANDS:
       help, h  Shows a list of commands or help for one command
    
    GLOBAL OPTIONS:
       --domain value       Set the website domain. (default: "sigsink.sslhound.com") [$DOMAIN]
       --enable-keyfetch    Enable fetching remote keys (default: false) [$ENABLE_KEYFETCH]
       --environment value  Set the environment the application is running in. (default: "development") [$ENVIRONMENT]
       --key-source value   A location to load keys from (default: "./keys") [$KEY_SOURCE]
       --listen value       Configure the server to listen to this interface. (default: "0.0.0.0:7000") [$LISTEN]
       --help, -h           show help (default: false)
       --version, -v        print the version (default: false)
    
    COPYRIGHT:
       (c) 2020 SSL Hound, LLC