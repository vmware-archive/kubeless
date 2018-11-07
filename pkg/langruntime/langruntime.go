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
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// PhaseInstallation - Installation phase name
	PhaseInstallation = "installation"
	// PhaseCompilation - Compilation phase name
	PhaseCompilation = "compilation"
	// PhaseRuntime - Runtime phase name
	PhaseRuntime = "runtime"
)

// Langruntimes struct for getting configmap
type Langruntimes struct {
	kubelessConfig    *v1.ConfigMap
	AvailableRuntimes []RuntimeInfo
}

// Image represents the information about a runtime phase
type Image struct {
	Phase   string            `yaml:"phase"`
	Image   string            `yaml:"image"`
	Command string            `yaml:"command,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// RuntimeVersion is a struct with all the info about the images and secrets
type RuntimeVersion struct {
	Name             string        `yaml:"name"`
	Version          string        `yaml:"version"`
	Images           []Image       `yaml:"runtimeImage"`
	ImagePullSecrets []ImageSecret `yaml:"imagePullSecrets,omitempty"`
}

// ImageSecret for pulling the image
type ImageSecret struct {
	ImageSecret string `yaml:"imageSecret,omitempty"`
}

// RuntimeInfo describe the runtime specifics (typical file suffix and dependency file name)
// and the supported versions
type RuntimeInfo struct {
	ID                string           `yaml:"ID"`
	Versions          []RuntimeVersion `yaml:"versions"`
	LivenessProbeInfo *v1.Probe        `yaml:"livenessProbeInfo,omitempty"`
	DepName           string           `yaml:"depName"`
	FileNameSuffix    string           `yaml:"fileNameSuffix"`
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
			if l.findImage(PhaseRuntime, l.AvailableRuntimes[i].Versions[j]) != nil {
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
	runtimeID := regexp.MustCompile("^[a-zA-Z_-]+").FindString(runtime)
	for _, runtimeInf := range l.AvailableRuntimes {
		if runtimeInf.ID == runtimeID {
			return runtimeInf, nil
		}
	}

	return RuntimeInfo{}, fmt.Errorf("Unable to find %s as runtime", runtime)
}

// GetLivenessProbeInfo returs the liveness probe info regarding a runtime
func (l *Langruntimes) GetLivenessProbeInfo(runtime string, port int) *v1.Probe {
	livenessProbe := &v1.Probe{
		InitialDelaySeconds: int32(3),
		PeriodSeconds:       int32(30),
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(port),
			},
		},
	}

	runtimeID := regexp.MustCompile("^[a-zA-Z]+").FindString(runtime)
	for _, runtimeInf := range l.AvailableRuntimes {
		if runtimeInf.ID == runtimeID {
			if runtimeInf.LivenessProbeInfo != nil {
				return runtimeInf.LivenessProbeInfo
			}
			return livenessProbe
		}
	}
	return livenessProbe
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

// Returns the image information of a phase or null if the phase is not found
func (l *Langruntimes) findImage(phase string, runtime RuntimeVersion) *Image {
	for _, imageInf := range runtime.Images {
		if imageInf.Phase == phase {
			return &imageInf
		}
	}
	return nil
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
		runtimeImage := l.findImage(PhaseRuntime, versionInf)
		if runtimeImage == nil {
			err = fmt.Errorf("The given runtime and version '%s' does not have a valid image for HTTP based functions. Available runtimes are: %s", runtime, strings.Join(l.getAvailableRuntimesPerTrigger("HTTP")[:], ", "))
		} else {
			imageName = runtimeImage.Image
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

func appendToCommand(orig string, command ...string) string {
	if len(orig) > 0 {
		return fmt.Sprintf("%s && %s", orig, strings.Join(command, " && "))
	}
	return strings.Join(command, " && ")
}

func parseEnv(env map[string]string) []v1.EnvVar {
	res := []v1.EnvVar{}
	for key, value := range env {
		res = append(res, v1.EnvVar{Name: key, Value: value})
	}
	return res
}

// GetBuildContainer returns a Container definition based on a runtime
func (l *Langruntimes) GetBuildContainer(runtime, depsChecksum string, env []v1.EnvVar, installVolume v1.VolumeMount) (v1.Container, error) {
	runtimeInf, err := l.GetRuntimeInfo(runtime)
	if err != nil {
		return v1.Container{}, err
	}
	depsFile := path.Join(installVolume.MountPath, runtimeInf.DepName)
	versionInf, err := l.findRuntimeVersion(runtime)
	if err != nil {
		return v1.Container{}, err
	}

	imageInf := l.findImage(PhaseInstallation, versionInf)
	if imageInf == nil {
		// The runtime doesn't have an installation hook
		return v1.Container{}, nil
	}

	var command string
	// Validate deps checksum
	shaFile := "/tmp/deps.sha256"
	command = appendToCommand(command,
		fmt.Sprintf("echo '%s  %s' > %s", depsChecksum, depsFile, shaFile),
		fmt.Sprintf("sha256sum -c %s", shaFile),
		imageInf.Command,
	)

	env = append(
		env,
		v1.EnvVar{Name: "KUBELESS_INSTALL_VOLUME", Value: installVolume.MountPath},
		v1.EnvVar{Name: "KUBELESS_DEPS_FILE", Value: depsFile},
	)
	env = append(env, parseEnv(imageInf.Env)...)

	return v1.Container{
		Name:            "install",
		Image:           imageInf.Image,
		Command:         []string{"sh", "-c"},
		Args:            []string{command},
		VolumeMounts:    []v1.VolumeMount{installVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		WorkingDir:      installVolume.MountPath,
		Env:             env,
	}, nil
}

// UpdateDeployment object in case of custom runtime
func (l *Langruntimes) UpdateDeployment(dpm *v1beta1.Deployment, volPath, runtime string) {
	versionInf, err := l.findRuntimeVersion(runtime)
	if err != nil {
		// Not found an image for the given runtime
		return
	}
	dpm.Spec.Template.Spec.Containers[0].Env = append(
		dpm.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{Name: "KUBELESS_INSTALL_VOLUME", Value: volPath},
	)

	imageInf := l.findImage(PhaseRuntime, versionInf)
	if imageInf == nil {
		// Not found an image for the given runtime
		return
	}
	dpm.Spec.Template.Spec.Containers[0].Env = append(
		dpm.Spec.Template.Spec.Containers[0].Env,
		parseEnv(imageInf.Env)...,
	)
}

// GetCompilationContainer returns a Container definition based on a runtime
func (l *Langruntimes) GetCompilationContainer(runtime, funcName string, installVolume v1.VolumeMount) (*v1.Container, error) {
	versionInf, err := l.findRuntimeVersion(runtime)
	if err != nil {
		return nil, err
	}

	imageInf := l.findImage(PhaseCompilation, versionInf)
	if imageInf == nil {
		// The runtime doesn't have a compilation hook
		return nil, nil
	}

	env := append(
		parseEnv(imageInf.Env),
		v1.EnvVar{Name: "KUBELESS_INSTALL_VOLUME", Value: installVolume.MountPath},
		v1.EnvVar{Name: "KUBELESS_FUNC_NAME", Value: funcName},
	)
	return &v1.Container{
		Name:            "compile",
		Image:           imageInf.Image,
		Command:         []string{"sh", "-c"},
		Args:            []string{imageInf.Command},
		Env:             env,
		VolumeMounts:    []v1.VolumeMount{installVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		WorkingDir:      installVolume.MountPath,
	}, nil
}
