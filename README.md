# Backup webdav

A tiny utility to compress, encrypt and upload data to a webdav server.

## Usage

For uploading files:

```bash
cat ./cats.jpg | backup-webdav "a/b/c" "cats.jpg.gz.encrypted"
```

For uploading folders:

```bash
tar -cvf - ./folder | ./backup-webdav "a/b/c" "folder.tar.gz.encrypted"
```
