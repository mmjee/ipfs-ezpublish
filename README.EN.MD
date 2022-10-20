A very simple tool to authenticate via Mutual TLS to an IPFS node, upload a directory and publish it on an IPNS node.

```shell
go build -ldflags "-s -w" -o ./ipfs-ezpublish .
./ipfs-ezpublish -cert-file Your.crt -private-key Your.key -shell-url "https://<your node>" -target dist -ipns-key <IPNS Key ID>
```

### Licence

Licensed under AGPL 3.0 only, see [LICENCE](LICENCE).