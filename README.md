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
// Run kubeless server
$ kubeless --master <master_node>

// Submit function
$ kubeless --master <master_node> function create <function_name> \
           --from-file <file> \
           --handler <handler> \
           --runtime <runtime>

// Delete function
$ kubeless --master <master_node> function delete <function_name>
```
