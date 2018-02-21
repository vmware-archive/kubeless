package langruntime

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

// Langruntimes struct for getting configmap
type Langruntimes struct {
	kubelessConfig    *v1.ConfigMap
	AvailableRuntimes []RuntimeInfo
}

// RuntimeVersion is a struct with all the info about the images and secrets
type RuntimeVersion struct {
	Name             string        `yaml:"name"`
	Version          string        `yaml:"version"`
	RuntimeImage     string        `yaml:"runtimeImage"`
	InitImage        string        `yaml:"initImage"`
	ImagePullSecrets []ImageSecret `yaml:"imagePullSecrets,omitempty"`
}

// ImageSecret for pulling the image
type ImageSecret struct {
	ImageSecret string `yaml:"imageSecret,omitempty"`
}

// RuntimeInfo describe the runtime specifics (typical file suffix and dependency file name)
// and the supported versions
type RuntimeInfo struct {
	ID             string           `yaml:"ID"`
	Versions       []RuntimeVersion `yaml:"versions"`
	DepName        string           `yaml:"depName"`
	FileNameSuffix string           `yaml:"fileNameSuffix"`
}

// New initializes a langruntime object
func New(config *v1.ConfigMap) *Langruntimes {
	var ri []RuntimeInfo

	return &Langruntimes{
		kubelessConfig:    config,
		AvailableRuntimes: ri,
	}
}

// ReadConfigMap reads the configmap
func (l *Langruntimes) ReadConfigMap() {
	if runtimeImages, ok := l.kubelessConfig.Data["runtime-images"]; ok {
		err := yaml.Unmarshal([]byte(runtimeImages), &l.AvailableRuntimes)
		if err != nil {
			logrus.Fatalf("Unable to get the runtime images: %v", err)
		}
	}
}

// GetRuntimes returns the list of available runtimes as strings
func (l *Langruntimes) GetRuntimes() []string {
	result := []string{}
	for _, runtimeInf := range l.AvailableRuntimes {
		for _, runtime := range runtimeInf.Versions {
			result = append(result, runtimeInf.ID+runtime.Version)
		}
	}
	return result
}

// IsValidRuntime returns true if passed runtime name is valid runtime
func (l *Langruntimes) IsValidRuntime(runtime string) bool {
	for _, validRuntime := range l.GetRuntimes() {
		if runtime == validRuntime {
			return true
		}
	}
	return false
}

func (l *Langruntimes) getAvailableRuntimesPerTrigger(imageType string) []string {
	var runtimeList []string
	for i := range l.AvailableRuntimes {
		for j := range l.AvailableRuntimes[i].Versions {
			if l.AvailableRuntimes[i].Versions[j].RuntimeImage != "" {
				runtimeList = append(runtimeList, l.AvailableRuntimes[i].ID+l.AvailableRuntimes[i].Versions[j].Version)
			}
		}
	}
	return runtimeList
}

// extract the branch number from the runtime string
func (l *Langruntimes) getVersionFromRuntime(runtime string) string {
	re := regexp.MustCompile("[0-9.]+$")
	return re.FindString(runtime)
}

// GetRuntimeInfo returns all the info regarding a runtime
func (l *Langruntimes) GetRuntimeInfo(runtime string) (RuntimeInfo, error) {
	runtimeID := regexp.MustCompile("^[a-zA-Z]+").FindString(runtime)
	for _, runtimeInf := range l.AvailableRuntimes {
		if runtimeInf.ID == runtimeID {
			return runtimeInf, nil
		}
	}
	return RuntimeInfo{}, fmt.Errorf("Unable to find %s as runtime", runtime)
}

func (l *Langruntimes) findRuntimeVersion(runtimeWithVersion string) (RuntimeVersion, error) {
	version := l.getVersionFromRuntime(runtimeWithVersion)
	runtimeInf, err := l.GetRuntimeInfo(runtimeWithVersion)
	if err != nil {
		return RuntimeVersion{}, err
	}
	for _, versionInf := range runtimeInf.Versions {
		if versionInf.Version == version {
			return versionInf, nil
		}
	}
	return RuntimeVersion{}, fmt.Errorf("The given runtime and version %s is not valid", runtimeWithVersion)
}

// GetFunctionImage returns the image ID depending on the runtime, its version and function type
func (l *Langruntimes) GetFunctionImage(runtime string) (string, error) {
	runtimeInf, err := l.GetRuntimeInfo(runtime)
	if err != nil {
		return "", err
	}

	imageNameEnvVar := strings.ToUpper(runtimeInf.ID) + l.getVersionFromRuntime(runtime) + "_RUNTIME"
	imageName := os.Getenv(imageNameEnvVar)
	if imageName == "" {
		versionInf, err := l.findRuntimeVersion(runtime)
		if err != nil {
			return "", err
		}
		if versionInf.RuntimeImage == "" {
			err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for HTTP based functions. Available runtimes are: %s", runtime, strings.Join(l.getAvailableRuntimesPerTrigger("HTTP")[:], ", "))
		} else {
			imageName = versionInf.RuntimeImage
		}
	}
	return imageName, nil
}

// GetImageSecrets gets the secrets to pull the runtime image
func (l *Langruntimes) GetImageSecrets(runtime string) ([]v1.LocalObjectReference, error) {
	var secrets []string

	runtimeInf, err := l.findRuntimeVersion(runtime)
	if err != nil {
		return []v1.LocalObjectReference{}, err
	}

	if len(runtimeInf.ImagePullSecrets) == 0 {
		return []v1.LocalObjectReference{}, nil
	}

	for _, s := range runtimeInf.ImagePullSecrets {
		secrets = append(secrets, s.ImageSecret)
	}
	var lors []v1.LocalObjectReference
	if len(secrets) > 0 {
		for _, s := range secrets {
			lor := v1.LocalObjectReference{Name: s}
			lors = append(lors, lor)
		}
	}

	return lors, nil
}

// GetBuildContainer returns a Container definition based on a runtime
func (l *Langruntimes) GetBuildContainer(runtime string, env []v1.EnvVar, installVolume v1.VolumeMount) (v1.Container, error) {
	runtimeInf, err := l.GetRuntimeInfo(runtime)
	if err != nil {
		return v1.Container{}, err
	}
	depsFile := path.Join(installVolume.MountPath, runtimeInf.DepName)
	versionInf, err := l.findRuntimeVersion(runtime)
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

	case strings.Contains(runtime, "php"):
		command = "composer install -d " + installVolume.MountPath
	}

	return v1.Container{
		Name:            "install",
		Image:           versionInf.InitImage,
		Command:         []string{"sh", "-c"},
		Args:            []string{command},
		VolumeMounts:    []v1.VolumeMount{installVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		WorkingDir:      installVolume.MountPath,
		Env:             env,
	}, nil
}

// UpdateDeployment object in case of custom runtime
func (l *Langruntimes) UpdateDeployment(dpm *v1beta1.Deployment, depsPath, runtime string) {
	switch {
	case strings.Contains(runtime, "python"):
		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "PYTHONPATH",
			Value: path.Join(depsPath, "lib/python"+l.getVersionFromRuntime(runtime)+"/site-packages"),
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
