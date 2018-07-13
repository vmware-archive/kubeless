/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"strconv"
	"time"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EnsureCronJob creates/updates a function cron job
func EnsureCronJob(client kubernetes.Interface, funcObj *kubelessApi.Function, schedule, reqImage string, or []metav1.OwnerReference, reqImagePullSecret []v1.LocalObjectReference) error {
	var maxSucccessfulHist, maxFailedHist int32
	maxSucccessfulHist = 3
	maxFailedHist = 1
	var timeout int
	if funcObj.Spec.Timeout != "" {
		var err error
		timeout, err = strconv.Atoi(funcObj.Spec.Timeout)
		if err != nil {
			return fmt.Errorf("Unable convert %s to a valid timeout", funcObj.Spec.Timeout)
		}
	} else {
		timeout, _ = strconv.Atoi(defaultTimeout)
	}
	activeDeadlineSeconds := int64(timeout)
	jobName := fmt.Sprintf("trigger-%s", funcObj.ObjectMeta.Name)
	var headersString = ""
	timestamp := time.Now().UTC()
	eventID, err := GetRandString(11)
	if err != nil {
		return fmt.Errorf("Failed to create a event-ID %v", err)
	}
	headersString = headersString + " -H \"event-id: " + eventID + "\""
	headersString = headersString + " -H \"event-time: " + timestamp.String() + "\""
	headersString = headersString + " -H \"event-type: application/json\""
	headersString = headersString + " -H \"event-namespace: cronjobtrigger.kubeless.io\""
	job := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       funcObj.ObjectMeta.Namespace,
			Labels:          addDefaultLabel(funcObj.ObjectMeta.Labels),
			OwnerReferences: or,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                   schedule,
			SuccessfulJobsHistoryLimit: &maxSucccessfulHist,
			FailedJobsHistoryLimit:     &maxFailedHist,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds: &activeDeadlineSeconds,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							ImagePullSecrets: reqImagePullSecret,
							Containers: []v1.Container{
								{
									Image: reqImage,
									Name:  "trigger",
									Args:  []string{"curl", "-Lv", headersString, fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", funcObj.ObjectMeta.Name, funcObj.ObjectMeta.Namespace)},
								},
							},
							RestartPolicy: v1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	_, err = client.BatchV1beta1().CronJobs(funcObj.ObjectMeta.Namespace).Create(job)
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		newCronJob := &batchv1beta1.CronJob{}
		newCronJob, err = client.BatchV1beta1().CronJobs(funcObj.ObjectMeta.Namespace).Get(jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !hasDefaultLabel(newCronJob.ObjectMeta.Labels) {
			return fmt.Errorf("Found a conflicting cronjob object %s/%s. Aborting", funcObj.ObjectMeta.Namespace, funcObj.ObjectMeta.Name)
		}
		newCronJob.ObjectMeta.Labels = funcObj.ObjectMeta.Labels
		newCronJob.ObjectMeta.OwnerReferences = or
		newCronJob.Spec = job.Spec
		_, err = client.BatchV1beta1().CronJobs(funcObj.ObjectMeta.Namespace).Update(newCronJob)
	}
	return err
}

func addDefaultLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["created-by"] = "kubeless"
	return labels
}

func hasDefaultLabel(labels map[string]string) bool {
	if labels == nil || labels["created-by"] != "kubeless" {
		return false
	}
	return true
}
