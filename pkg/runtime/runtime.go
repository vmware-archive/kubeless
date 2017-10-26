package runtime

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	python27Http    = "bitnami/kubeless-python@sha256:6789266df0c97333f76e23efd58cf9c7efe24fa3e83b5fc826fd5cc317699b55"
	python27Pubsub  = "bitnami/kubeless-event-consumer@sha256:5ce469529811acf49c4d20bcd8a675be7aa029b43cf5252a8c9375b170859d83"
	python34Http    = "bitnami/kubeless-python:test@sha256:686cd28cda5fe7bc6db60fa3e8a9a2c57a5eff6a58e66a60179cc1d3fcf1035b"
	python34Pubsub  = "bitnami/kubeless-python-event-consumer@sha256:8f92397258836e9c39948814aa5324c29d96ff3624b66dd70fdbad1ce0a1615e"
	node6Http       = "bitnami/kubeless-nodejs@sha256:b3c7cec77f973bf7a48cbbb8ea5069cacbaee7044683a275c6f78fa248de17b4"
	node6Pubsub     = "bitnami/kubeless-nodejs-event-consumer@sha256:b027bfef5f99c3be68772155a1feaf1f771ab9a3c7bb49bef2e939d6b766abec"
	node8Http       = "bitnami/kubeless-nodejs@sha256:1eff2beae6fcc40577ada75624c3e4d3840a854588526cd8616d66f4e889dfe6"
	node8Pubsub     = "bitnami/kubeless-nodejs-event-consumer@sha256:4d005c9c0b462750d9ab7f1305897e7a01143fe869d3b722ed3330560f9c7fb5"
	ruby24Http      = "bitnami/kubeless-ruby@sha256:97b18ac36bb3aa9529231ea565b339ec00d2a5225cf7eb010cd5a6188cf72ab5"
	ruby24Pubsub    = "bitnami/kubeless-ruby-event-consumer@sha256:938a860dbd9b7fb6b4338248a02c92279315c6e42eed0700128b925d3696b606"
	dotnetcore2Http = "allantargino/kubeless-dotnetcore@sha256:d321dc4b2c420988d98cdaa22c733743e423f57d1153c89c2b99ff0d944e8a63"
	pubsubFunc      = "PubSub"
)

type runtimeVersion struct {
	runtimeID   string
	version     string
	httpImage   string
	pubsubImage string
}

var python, node, ruby, dotnetcore []runtimeVersion
var availableRuntimes [][]runtimeVersion

func init() {
	python27 := runtimeVersion{runtimeID: "python", version: "2.7", httpImage: python27Http, pubsubImage: python27Pubsub}
	python34 := runtimeVersion{runtimeID: "python", version: "3.4", httpImage: python34Http, pubsubImage: python34Pubsub}
	python = []runtimeVersion{python27, python34}

	node6 := runtimeVersion{runtimeID: "nodejs", version: "6", httpImage: node6Http, pubsubImage: node6Pubsub}
	node8 := runtimeVersion{runtimeID: "nodejs", version: "8", httpImage: node8Http, pubsubImage: node8Pubsub}
	node = []runtimeVersion{node6, node8}

	ruby24 := runtimeVersion{runtimeID: "ruby", version: "2.4", httpImage: ruby24Http, pubsubImage: ruby24Pubsub}
	ruby = []runtimeVersion{ruby24}

	dotnetcore2 := runtimeVersion{runtimeID: "dotnetcore", version: "2.0", httpImage: dotnetcore2Http, pubsubImage: ""}
	dotnetcore = []runtimeVersion{dotnetcore2}

	availableRuntimes = [][]runtimeVersion{python, node, ruby, dotnetcore}
}

// GetRuntimes returns the list of available runtimes as strings
func GetRuntimes() []string {
	result := []string{}
	for _, languages := range availableRuntimes {
		for _, runtime := range languages {
			result = append(result, runtime.runtimeID+runtime.version)
		}
	}
	return result
}

func getAvailableRuntimes(imageType string) []string {
	runtimeObjList := [][]runtimeVersion{python, node, ruby}
	var runtimeList []string
	for i := range runtimeObjList {
		for j := range runtimeObjList[i] {
			if (imageType == "PubSub" && runtimeObjList[i][j].pubsubImage != "") || (imageType == "HTTP" && runtimeObjList[i][j].httpImage != "") {
				runtimeList = append(runtimeList, runtimeObjList[i][j].runtimeID+runtimeObjList[i][j].version)
			}
		}
	}
	return runtimeList
}

// GetRuntimeDepName returns the common dependency file name of the given runtime
func GetRuntimeDepName(runtime string) (string, error) {
	depName := ""
	switch {
	case strings.Contains(runtime, "python"):
		depName = "requirements.txt"
	case strings.Contains(runtime, "nodejs"):
		depName = "package.json"
	case strings.Contains(runtime, "ruby"):
		depName = "Gemfile"
	case strings.Contains(runtime, "dotnetcore"):
		depName = "requirements.xml"
	default:
		return "", errors.New("The given runtime is not valid")
	}
	return depName, nil
}

// GetFunctionFileName returns the file name given a handler ID and a runtime
func GetFunctionFileName(modName, runtime string) string {
	fileName := ""
	switch {
	case strings.Contains(runtime, "python"):
		fileName = modName + ".py"
	case strings.Contains(runtime, "nodejs"):
		fileName = modName + ".js"
	case strings.Contains(runtime, "ruby"):
		fileName = modName + ".rb"
	case strings.Contains(runtime, "dotnetcore"):
		fileName = modName + ".cs"
	default:
		fileName = modName
	}
	return fileName
}

// GetFunctionImage returns the image ID depending on the runtime, its version and function type
func GetFunctionImage(runtime, ftype string) (imageName string, err error) {
	runtimeID := regexp.MustCompile("[a-zA-Z]+").FindString(runtime)
	version := regexp.MustCompile("[0-9.]+").FindString(runtime)
	var versionsDef []runtimeVersion
	var httpImage, pubsubImage string
	switch {
	case runtimeID == "python":
		versionsDef = python
	case runtimeID == "nodejs":
		versionsDef = node
	case runtimeID == "ruby":
		versionsDef = ruby
	case runtimeID == "dotnetcore":
		versionsDef = dotnetcore
	default:
		err = errors.New("The given runtime is not valid")
		return
	}

	for i := range versionsDef {
		if versionsDef[i].version == version {
			httpImage = versionsDef[i].httpImage
			pubsubImage = versionsDef[i].pubsubImage
		}
	}

	imageNameEnvVar := ""
	if ftype == pubsubFunc {
		imageNameEnvVar = strings.ToUpper(runtimeID) + "_PUBSUB_RUNTIME"
	} else {
		imageNameEnvVar = strings.ToUpper(runtimeID) + "_RUNTIME"
	}
	if imageName = os.Getenv(imageNameEnvVar); imageName == "" {
		if ftype == pubsubFunc {
			if pubsubImage == "" {
				err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for event based functions. Available runtimes are: %s", runtime, strings.Join(getAvailableRuntimes("PubSub")[:], ", "))
			} else {
				imageName = pubsubImage
			}
		} else {
			if httpImage == "" {
				err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for HTTP based functions. Available runtimes are: %s", runtime, strings.Join(getAvailableRuntimes("HTTP")[:], ", "))
			} else {
				imageName = httpImage
			}
		}
	}
	return
}

// extract the branch number from the runtime string
func getBranchFromRuntime(runtime string) string {
	re := regexp.MustCompile("[0-9.]+")
	return re.FindString(runtime)
}

// specify image for the init container
func getInitImagebyRuntime(runtime string) string {
	switch {
	case strings.Contains(runtime, "python"):
		branch := getBranchFromRuntime(runtime)
		if branch == "2.7" {
			// TODO: Migrate the image for python 2.7 to an official source (not alpine-based)
			return "tuna/python-pillow:2.7.11-alpine"
		}
		return "python:" + branch
	case strings.Contains(runtime, "nodejs"):
		return "node:6.10"
	case strings.Contains(runtime, "ruby"):
		return "bitnami/ruby:2.4"
	case strings.Contains(runtime, "dotnetcore"):
		return "microsoft/aspnetcore-build:2.0"
	default:
		return ""
	}
}

// GetBuildContainer returns a Container definition based on a runtime
func GetBuildContainer(runtime string, env []v1.EnvVar, runtimeVolume, depsVolume v1.VolumeMount) (v1.Container, error) {
	depName, err := GetRuntimeDepName(runtime)
	if err != nil {
		return v1.Container{}, err
	}
	depsFile := path.Join(depsVolume.MountPath, depName)
	var command string
	switch {
	case strings.Contains(runtime, "python"):
		command = "pip install --prefix=" + runtimeVolume.MountPath + " -r " + depsFile
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
			" && npm install " + depsVolume.MountPath + " --prefix=" + runtimeVolume.MountPath
	case strings.Contains(runtime, "ruby"):
		command = "bundle install --gemfile=" + depsFile + " --path=" + runtimeVolume.MountPath
	}
	return v1.Container{
		Name:            "install",
		Image:           getInitImagebyRuntime(runtime),
		Command:         []string{"sh", "-c"},
		Args:            []string{command},
		VolumeMounts:    []v1.VolumeMount{runtimeVolume, depsVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		Env:             env,
	}, nil
}

// UpdateDeployment object in case of custom runtime
func UpdateDeployment(dpm *v1beta1.Deployment, depsVolume, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "PYTHONPATH",
			Value: "/opt/kubeless/pythonpath/lib/python" + getBranchFromRuntime(runtime) + "/site-packages",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      depsVolume,
			MountPath: "/opt/kubeless/pythonpath",
		})
	case strings.Contains(runtime, "nodejs"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "NODE_PATH",
			Value: "/opt/kubeless/nodepath/node_modules",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      depsVolume,
			MountPath: "/opt/kubeless/nodepath",
		})
	case strings.Contains(runtime, "ruby"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "GEM_HOME",
			Value: "/opt/kubeless/rubypath/ruby/2.4.0",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      depsVolume,
			MountPath: "/opt/kubeless/rubypath",
		})
	case strings.Contains(runtime, "dotnetcore"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "DOTNETCORE_HOME",
			Value: "/usr/bin/",
		})
		dpm.Spec.Template.Spec.Containers[0].VolumeMounts = append(dpm.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      depsVolume,
			MountPath: "/opt/kubeless/dotnetcorepath",
		})
	}
}
