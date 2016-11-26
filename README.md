# Kubeless
[![Join us on Slack](https://s3.eu-central-1.amazonaws.com/ngtuna/join-us-on-slack.png)](https://skippbox.herokuapp.com)

## Building

### Building with go

- you need go v1.5 or later.
- if your working copy is not in your `GOPATH`, you need to set it accordingly.

```console
$ go build -o kubeless main.go
```

## Download kubeless package

```console
$ go get -u github.com/skippbox/kubeless
```

### Run kubeless

**Kubectl is required**

```console
// Proxy kubectl
$ kubectl proxy -p 8080

// Install Kubeless controller
$ kubeless install

// Submit function
$ kubeless function create <function_name> \
           --from-file <file> \
           --handler <handler> \
           --runtime <runtime>

// Delete function
$ kubeless function delete <function_name>

// List running functions
$ kubeless function ls
```
