# What?

bserv - Simple backup server which stores uploaded files. A subdirectory with current date will be created automatically.

# REST API

`PUT /up?name=$NAME`

Where: `$NAME` - Any file name;

Request body: File content.

Response: Status 200 OK and empty body.

# Requirements

Go >= 1.16.

# Usage

Server arguments:

```
bserv [-listen-on <IP:PORT>] <DIR>
-listen-on <IP:PORT> - Listen on the IP and PORT (:2180 by default);
<DIR> - Path to the root directory;
```

Uploading file:

`curl --upload-file readme.txt 'http://example.com/up?name=readme.txt'`

Or from stdin:

`cat readme.txt | curl --upload-file - 'http://example.com/up?name=readme.txt'`

# License

GPL.
