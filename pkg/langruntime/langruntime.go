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
	ID                string           `yaml:"ID"`
	Compiled          bool             `yaml:"compiled"`
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
			} else {
				return livenessProbe
			}
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

func appendToCommand(orig string, command ...string) string {
	if len(orig) > 0 {
		return fmt.Sprintf("%s && %s", orig, strings.Join(command, " && "))
	}
	return strings.Join(command, " && ")
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

	var command string
	// Validate deps checksum
	shaFile := "/tmp/deps.sha256"
	command = appendToCommand(command,
		fmt.Sprintf("echo '%s  %s' > %s", depsChecksum, depsFile, shaFile),
		fmt.Sprintf("sha256sum -c %s", shaFile))

	switch {
	case strings.Contains(runtime, "python"):
		command = appendToCommand(command,
			"pip install --prefix="+installVolume.MountPath+" -r "+depsFile)
	case strings.Contains(runtime, "nodejs"):
		registry := "https://registry.npmjs.org"
		scope := ""
		// Force HOME to a folder with permissions to avoid issues in OpenShift #694
		env = append(env, v1.EnvVar{Name: "HOME", Value: "/tmp"})
		for _, v := range env {
			if v.Name == "NPM_REGISTRY" {
				registry = v.Value
			}
			if v.Name == "NPM_SCOPE" {
				scope = v.Value + ":"
			}
		}
		command = appendToCommand(command,
			"npm config set "+scope+"registry "+registry,
			"npm install --production --prefix="+installVolume.MountPath)
	case strings.Contains(runtime, "ruby"):
		command = appendToCommand(command,
			"bundle install --gemfile="+depsFile+" --path="+installVolume.MountPath)

	case strings.Contains(runtime, "php"):
		command = appendToCommand(command,
			"composer install -d "+installVolume.MountPath)
	case strings.Contains(runtime, "go"):
		command = appendToCommand(command,
			"cd $GOPATH/src/kubeless",
			"dep ensure > /dev/termination-log 2>&1")
	case strings.Contains(runtime, "dotnetcore"):
		logrus.Warn("dotnetcore does not require a dependencies file")
		return v1.Container{}, nil
	case strings.Contains(runtime, "java"):
		command = appendToCommand(command,
			"mv /kubeless/pom.xml /kubeless/function-pom.xml")
	case strings.Contains(runtime, "ballerina"):
		return v1.Container{}, fmt.Errorf("Ballerina does not require a dependencies file")
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
			Value: path.Join(depsPath, "packages"),
		})
	}
}

// RequiresCompilation returns if the given runtime requires compilation
func (l *Langruntimes) RequiresCompilation(runtime string) bool {
	required := false
	for _, runtimeInf := range l.AvailableRuntimes {
		if strings.Contains(runtime, runtimeInf.ID) {
			required = runtimeInf.Compiled
			break
		}
	}
	return required
}

// GetCompilationContainer returns a Container definition based on a runtime
func (l *Langruntimes) GetCompilationContainer(runtime, funcName string, installVolume v1.VolumeMount) (v1.Container, error) {
	versionInf, err := l.findRuntimeVersion(runtime)
	if err != nil {
		return v1.Container{}, err
	}
	var command string
	switch {
	case strings.Contains(runtime, "go"):
		command = fmt.Sprintf(
			"sed 's/<<FUNCTION>>/%s/g' $GOPATH/src/controller/kubeless.tpl.go > $GOPATH/src/controller/kubeless.go && "+
				"go build -o %s/server $GOPATH/src/controller/kubeless.go > /dev/termination-log 2>&1", funcName, installVolume.MountPath)
	case strings.Contains(runtime, "java"):
		command = "cp -r /usr/src/myapp/* /kubeless/ && " +
			"cp /kubeless/*.java /kubeless/function/src/main/java/io/kubeless/ && " +
			"cp /kubeless/function-pom.xml /kubeless/function/pom.xml 2>/dev/null || true && " +
			"mvn package > /dev/termination-log 2>&1 && mvn install > /dev/termination-log 2>&1"
	case strings.Contains(runtime, "dotnetcore"):
		command = "/app/compile-function.sh " + installVolume.MountPath
	case strings.Contains(runtime, "ballerina"):
		command = fmt.Sprintf("/compile-function.sh %s", funcName)

	default:
		return v1.Container{}, fmt.Errorf("Not found a valid compilation step for %s", runtime)
	}
	return v1.Container{
		Name:            "compile",
		Image:           versionInf.InitImage,
		Command:         []string{"sh", "-c"},
		Args:            []string{command},
		VolumeMounts:    []v1.VolumeMount{installVolume},
		ImagePullPolicy: v1.PullIfNotPresent,
		WorkingDir:      installVolume.MountPath,
	}, nil
}
