package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

var usage = "Usage: ksonnet-gen [path to k8s OpenAPI swagger.json] [output dir]"

func main() {
	if len(os.Args) != 3 {
		log.Fatal(usage)
	}

	swaggerPath := os.Args[1]
	text, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		log.Fatalf("Could not read file at '%s':\n%v", swaggerPath, err)
	}

	// Deserialize the API object.
	s := kubespec.APISpec{}
	err = json.Unmarshal(text, &s)
	if err != nil {
		log.Fatalf("Could not deserialize schema:\n%v", err)
	}
	s.Text = text
	s.FilePath = filepath.Dir(swaggerPath)

	// Emit Jsonnet code.
	jsonnetBytes, err := ksonnet.Emit(&s)
	if err != nil {
		log.Fatalf("Could not write ksonnet library:\n%v", err)
	}

	// Write out.
	outfile := fmt.Sprintf("%s/%s", os.Args[2], "k8s.libsonnet")
	err = ioutil.WriteFile(outfile, jsonnetBytes, 0644)
	if err != nil {
		log.Fatalf("Could not write `kube.libsonnet`:\n%v", err)
	}
}

func init() {
	// Get rid of time in logs.
	log.SetFlags(0)
}
