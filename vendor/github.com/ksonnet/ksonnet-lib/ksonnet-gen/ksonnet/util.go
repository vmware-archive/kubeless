package ksonnet

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

const constructorName = "new"

// isMixinRef will check whether a `ObjectRef` refers to an API object
// that can be turned into a mixin. This should be true of the vast
// majority of non-nil `ObjectRef`s. The most common exception is
// `IntOrString`, which should not be turned into a mixin, and should
// instead by transformed into a property method that behaves
// identically to one taking an int or a ref as argument.
func isMixinRef(or *kubespec.ObjectRef) bool {
	return or != nil && *or != "#/definitions/io.k8s.apimachinery.pkg.util.intstr.IntOrString"
}

var specialProperties = map[kubespec.PropertyName]kubespec.PropertyName{
	"apiVersion": "apiVersion",
	"kind":       "kind",
}

var specialPropertiesList = []string{"apiVersion", "kind"}

func isSpecialProperty(pn kubespec.PropertyName) bool {
	_, ok := specialProperties[pn]
	return ok
}

func getSHARevision(dir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could get working directory:\n%v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		log.Fatalf("Could cd to directory of repository at '%s':\n%v", dir, err)
	}

	sha, err := exec.Command("sh", "-c", "git rev-parse HEAD").Output()
	if err != nil {
		log.Fatalf("Could not find SHA of HEAD:\n%v", err)
	}

	err = os.Chdir(cwd)
	if err != nil {
		log.Fatalf("Could cd back to current directory '%s':\n%v", cwd, err)
	}

	return strings.TrimSpace(string(sha))
}
