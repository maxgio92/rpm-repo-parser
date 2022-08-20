# RPM repository parser

> Under development

From `mirrorURL` and `repoURI`:

```go
mirrorURL = "https://mirrors.edge.kernel.org/centos"
repoURI = "/8-stream/BaseOS/x86_64/os"
```

returns the list of provided packages with related metadata.

## Run a demo

```
go run main.go
```
