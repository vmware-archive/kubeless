package minio

import (
	"fmt"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

func waitForCompletedJob(jobName, namespace string, timeout int, cli kubernetes.Interface) error {
	counter := 0
	for counter < timeout {
		j, err := cli.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if j.Status.Failed == 1 {
			err = fmt.Errorf("Unable to run upload job. Received: %s", j.Status.Conditions[0].Message)
			return err
		} else if j.Status.Succeeded == 1 {
			return nil
		}
		time.Sleep(time.Duration(time.Second))
		counter++
	}
	return fmt.Errorf("Upload job has not finished after %s seconds", string(timeout))
}

// UploadFunction uploads a file to Minio using as object name file.extension
// It uses a Kubernetes job to access Minio since we need to use an URL <domain>[:port] (a URL with
// proxy is not valid)
func UploadFunction(file, checksum string, cli kubernetes.Interface) (string, error) {
	minioCredentials := "minio-key"
	jobName := "upload-file"
	fileName := path.Base(file) + "." + checksum
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "kubeless",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: minioCredentials,
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: minioCredentials,
								},
							},
						},
						{
							Name: "func",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: file,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:  "uploader",
							Image: "minio/mc:RELEASE.2017-10-14T00-51-16Z",
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      minioCredentials,
									MountPath: "/minio-cred",
								},
								{
									Name:      "func",
									MountPath: "/" + path.Base(file),
								},
							},
							Command: []string{"sh", "-c"},
							Args: []string{
								"mc config host add minioserver http://minio.kubeless:9000 $(cat /minio-cred/accesskey) $(cat /minio-cred/secretkey); " +
									"mc cp /" + path.Base(file) + " minioserver/functions/" + fileName,
							},
						},
					},
				},
			},
		},
	}
	_, err := cli.BatchV1().Jobs("kubeless").Create(&job)
	if err != nil {
		return "", err
	}
	err = waitForCompletedJob(jobName, "kubeless", 120, cli)
	if err != nil {
		return "", err
	}
	// Clean up (delete job)
	err = cli.BatchV1().Jobs("kubeless").Delete(jobName, &metav1.DeleteOptions{})
	if err != nil {
		return "", err
	}
	return "http://minio.kubeless:9000/functions/" + fileName, nil
}
