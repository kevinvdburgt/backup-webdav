# Backup WebDAV

A tiny utility to compress, encrypt and upload data to a WebDAV server.

## Usage

For uploading files:

```bash
cat ./cats.jpg | ./backup-webdav "a/b/c" "cats.jpg.gz.encrypted"
```

For uploading folders:

```bash
tar -cvf - ./folder | ./backup-webdav "a/b/c" "folder.tar.gz.encrypted"
```

## Notes

I've been using this personally with Hetzner Storage Box to create backups. Both on my local machine and on remote servers. It's been handy for simple, streamable backups over WebDAV.

This project is part of my journey learning Go. The code works reliably, but there’s probably plenty of room for improvements and refactoring. Feel free to open an issue or PR if you spot something worth tweaking!
