# Webmail

## Build

Requires Go 1.26 and Node 22.

```sh
cd web && npm install && npm run build && cd ..
go build -o webmail .
```

## Run

```sh
./webmail
```

It listens on 127.0.0.1:8080 and keeps its data in ./data. Two flags change that:

- `-addr` listen address
- `-data` data directory
