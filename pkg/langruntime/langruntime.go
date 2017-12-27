package langruntime

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	python27Http    = "kubeless/python@sha256:0f3b64b654df5326198e481cd26e73ecccd905aae60810fc9baea4dcbb61f697"
	python27Pubsub  = "kubeless/python-event-consumer@sha256:1aeb6cef151222201abed6406694081db26fa2235d7ac128113dcebd8d73a6cb"
	python27Init    = "tuna/python-pillow:2.7.11-alpine" // TODO: Migrate the image for python 2.7 to an official source (not alpine-based)
	python34Http    = "kubeless/python@sha256:e502078dc9580bb73f823504a6765dfc98f000979445cdf071900350b938c292"
	python34Pubsub  = "kubeless/python-event-consumer@sha256:d963e4cd58229d662188d618cd87503b3c749b126b359ce724a19a375e4b3040"
	python34Init    = "python:3.4"
	python36Http    = "kubeless/python@sha256:6300c2513ca51653ae698a31eacf6b2b8a16d2737dd3e244a8c9c11f6408fd35"
	python36Pubsub  = "kubeless/python-event-consumer@sha256:0a2f9162de56b7966b02b70a5a0bcff03badfd9d87b8ae3d13e5381abd00220f"
	python36Init    = "python:3.6"
	node6Http       = "kubeless/nodejs@sha256:2b25d7380d6ed06ad817f4ee1e177340a282788596b34464173bb8a967d83c02"
	node6Pubsub     = "kubeless/nodejs-event-consumer@sha256:1861c32d6a46b2fdfc3e3996daf690ff2c3d5ca19a605abd2af503011d68e221"
	node6Init       = "node:6.10"
	node8Http       = "kubeless/nodejs@sha256:f1426efe274ea8480d95270c98f6007ac64645e36291dbfa36d759b5c8b7b733"
	node8Pubsub     = "kubeless/nodejs-event-consumer@sha256:b301b02e463b586d9a32d5c1cb5a68c2a11e4fba9514e28d900fc50a78759af9"
	node8Init       = "node:8"
	ruby24Http      = "kubeless/ruby@sha256:738e4cdeb5f5feece236bbf4e46902024e4b9fc16db4f3791404fa27e8b0db15"
	ruby24Pubsub    = "kubeless/ruby-event-consumer@sha256:f9f50be51d93a98ae30689d87b067c181905a8757d339fb0fa9a81c6268c4eea"
	ruby24Init      = "bitnami/ruby:2.4"
	dotnetcore2Http = "allantargino/kubeless-dotnetcore@sha256:d321dc4b2c420988d98cdaa22c733743e423f57d1153c89c2b99ff0d944e8a63"
	dotnetcore2Init = "microsoft/aspnetcore-build:2.0"
	pubsubFunc      = "PubSub"
)

type runtimeVersion struct {
	version     string
	httpImage   string
	pubsubImage string
	initImage   string
}

// RuntimeInfo describe the runtime specifics (typical file suffix and dependency file name)
// and the supported versions
type RuntimeInfo struct {
	ID             string
	versions       []runtimeVersion
	DepName        string
	FileNameSuffix string
}

var pythonVersions, nodeVersions, rubyVersions, dotnetcoreVersions []runtimeVersion
var availableRuntimes []RuntimeInfo

func init() {
	python27 := runtimeVersion{version: "2.7", httpImage: python27Http, pubsubImage: python27Pubsub, initImage: python27Init}
	python34 := runtimeVersion{version: "3.4", httpImage: python34Http, pubsubImage: python34Pubsub, initImage: python34Init}
	python36 := runtimeVersion{version: "3.6", httpImage: python36Http, pubsubImage: python36Pubsub, initImage: python36Init}
	pythonVersions = []runtimeVersion{python27, python34, python36}

	node6 := runtimeVersion{version: "6", httpImage: node6Http, pubsubImage: node6Pubsub, initImage: node6Init}
	node8 := runtimeVersion{version: "8", httpImage: node8Http, pubsubImage: node8Pubsub, initImage: node8Init}
	nodeVersions = []runtimeVersion{node6, node8}

	ruby24 := runtimeVersion{version: "2.4", httpImage: ruby24Http, pubsubImage: ruby24Pubsub, initImage: ruby24Init}
	rubyVersions = []runtimeVersion{ruby24}

	dotnetcore2 := runtimeVersion{version: "2.0", httpImage: dotnetcore2Http, pubsubImage: "", initImage: dotnetcore2Init}
	dotnetcoreVersions = []runtimeVersion{dotnetcore2}

	availableRuntimes = []RuntimeInfo{
		{ID: "python", versions: pythonVersions, DepName: "requirements.txt", FileNameSuffix: ".py"},
		{ID: "nodejs", versions: nodeVersions, DepName: "package.json", FileNameSuffix: ".js"},
		{ID: "ruby", versions: rubyVersions, DepName: "Gemfile", FileNameSuffix: ".rb"},
		{ID: "dotnetcore", versions: dotnetcoreVersions, DepName: "requirements.xml", FileNameSuffix: ".cs"},
	}
}

// GetRuntimes returns the list of available runtimes as strings
func GetRuntimes() []string {
	result := []string{}
	for _, runtimeInf := range availableRuntimes {
		for _, runtime := range runtimeInf.versions {
			result = append(result, runtimeInf.ID+runtime.version)
		}
	}
	return result
}

// IsValidRuntime returns true if passed runtime name is valid runtime
func IsValidRuntime(runtime string) bool {
	for _, validRuntime := range GetRuntimes() {
		if runtime == validRuntime {
			return true
		}
	}
	return false
}

func getAvailableRuntimesPerTrigger(imageType string) []string {
	var runtimeList []string
	for i := range availableRuntimes {
		for j := range availableRuntimes[i].versions {
			if (imageType == "PubSub" && availableRuntimes[i].versions[j].pubsubImage != "") || (imageType == "HTTP" && availableRuntimes[i].versions[j].httpImage != "") {
				runtimeList = append(runtimeList, availableRuntimes[i].ID+availableRuntimes[i].versions[j].version)
			}
		}
	}
	return runtimeList
}

// extract the branch number from the runtime string
func getVersionFromRuntime(runtime string) string {
	re := regexp.MustCompile("[0-9.]+$")
	return re.FindString(runtime)
}

// GetRuntimeInfo returns all the info regarding a runtime
func GetRuntimeInfo(runtime string) (RuntimeInfo, error) {
	runtimeID := regexp.MustCompile("^[a-zA-Z]+").FindString(runtime)
	for _, runtimeInf := range availableRuntimes {
		if runtimeInf.ID == runtimeID {
			return runtimeInf, nil
		}
	}
	return RuntimeInfo{}, fmt.Errorf("Unable to find %s as runtime", runtime)
}

func findRuntimeVersion(runtimeWithVersion string) (runtimeVersion, error) {
	version := getVersionFromRuntime(runtimeWithVersion)
	runtimeInf, err := GetRuntimeInfo(runtimeWithVersion)
	if err != nil {
		return runtimeVersion{}, err
	}
	for _, versionInf := range runtimeInf.versions {
		if versionInf.version == version {
			return versionInf, nil
		}
	}
	return runtimeVersion{}, fmt.Errorf("The given runtime and version %s is not valid", runtimeWithVersion)
}

// GetFunctionImage returns the image ID depending on the runtime, its version and function type
func GetFunctionImage(runtime, ftype string) (string, error) {
	runtimeInf, err := GetRuntimeInfo(runtime)
	if err != nil {
		return "", err
	}

	imageNameEnvVar := ""
	if ftype == pubsubFunc {
		imageNameEnvVar = strings.ToUpper(runtimeInf.ID) + getVersionFromRuntime(runtime) + "_PUBSUB_RUNTIME"
	} else {
		imageNameEnvVar = strings.ToUpper(runtimeInf.ID) + getVersionFromRuntime(runtime) + "_RUNTIME"
	}
	imageName := os.Getenv(imageNameEnvVar)
	if imageName == "" {
		versionInf, err := findRuntimeVersion(runtime)
		if err != nil {
			return "", err
		}
		if ftype == pubsubFunc {
			if versionInf.pubsubImage == "" {
				err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for event based functions. Available runtimes are: %s", runtime, strings.Join(getAvailableRuntimesPerTrigger("PubSub")[:], ", "))
			} else {
				imageName = versionInf.pubsubImage
			}
		} else {
			if versionInf.httpImage == "" {
				err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for HTTP based functions. Available runtimes are: %s", runtime, strings.Join(getAvailableRuntimesPerTrigger("HTTP")[:], ", "))
			} else {
				imageName = versionInf.httpImage
			}
		}
	}
	return imageName, nil
}

// GetBuildContainer returns a Container definition based on a runtime
func GetBuildContainer(runtime string, env []v1.EnvVar, installVolume v1.VolumeMount) (v1.Container, error) {
	runtimeInf, err := GetRuntimeInfo(runtime)
	if err != nil {
		return v1.Container{}, err
	}
	depsFile := path.Join(installVolume.MountPath, runtimeInf.DepName)
	versionInf, err := findRuntimeVersion(runtime)
	if err != nil {
		return v1.Container{}, err
	}

	var command string
	switch {
	case strings.Contains(runtime, "python"):
		command = "pip install --prefix=" + installVolume.MountPath + " -r " + depsFile
	case strings.Contains(runtime, "nodejs"):
		registry := "https://registry.npmjs.org"
		scope := ""
		for _, v := range env {
			if v.Name == "NPM_REGISTRY" {
				registry = v.Value
			}
			if v.Name == "NPM_SCOPE" {
				scope = v.Value + ":"
			}
		}
		command = "npm config set " + scope + "registry " + registry +
			" && npm install --prefix=" + installVolume.MountPath
	case strings.Contains(runtime, "ruby"):
		command = "bundle install --gemfile=" + depsFile + " --path=" + installVolume.MountPath
	}

	return v1.Container{
		Name:            "install",
		Image:           versionInf.initImage,
		Command:         []string{"sh", "-c"},
		Args:            []string{command},
		VolumeMounts:    []v1.VolumeMount{installVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		WorkingDir:      installVolume.MountPath,
		Env:             env,
	}, nil
}

// UpdateDeployment object in case of custom runtime
func UpdateDeployment(dpm *v1beta1.Deployment, depsPath, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "PYTHONPATH",
			Value: path.Join(depsPath, "lib/python"+getVersionFromRuntime(runtime)+"/site-packages"),
		})
	case strings.Contains(runtime, "nodejs"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "NODE_PATH",
			Value: path.Join(depsPath, "node_modules"),
		})
	case strings.Contains(runtime, "ruby"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "GEM_HOME",
			Value: path.Join(depsPath, "ruby/2.4.0"),
		})
	case strings.Contains(runtime, "dotnetcore"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "DOTNETCORE_HOME",
			Value: "/usr/bin/",
		})
	}
}
