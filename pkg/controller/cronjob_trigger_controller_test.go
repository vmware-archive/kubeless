package controller

import (
	"testing"
	"time"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessFake "github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"github.com/sirupsen/logrus"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFunctionAddedUpdated(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	f := kubelessApi.Function{
		ObjectMeta: myNsFoo,
	}

	cjtrigger := kubelessApi.CronJobTrigger{
		ObjectMeta: myNsFoo,
	}

	triggerClientset := kubelessFake.NewSimpleClientset(&f, &cjtrigger)

	cronjob := batchv1beta1.CronJob{
		ObjectMeta: myNsFoo,
	}
	clientset := fake.NewSimpleClientset(&cronjob)

	controller := CronJobTriggerController{
		clientset:      clientset,
		kubelessclient: triggerClientset,
		logger:         logrus.WithField("controller", "cronjob-trigger-controller"),
	}

	// no-op for when the function is not deleted
	err := controller.functionAddedDeletedUpdated(&f, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	list, err := controller.kubelessclient.KubelessV1beta1().CronJobTriggers("myns").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].ObjectMeta.Name != "foo" {
		t.Errorf("Missing trigger in list: %v", list.Items)
	}
}

func TestFunctionDeleted(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	f := kubelessApi.Function{
		ObjectMeta: myNsFoo,
	}

	cjtrigger := kubelessApi.CronJobTrigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "myns",
			Name:      "foo-trigger",
		},
		Spec: kubelessApi.CronJobTriggerSpec{
			FunctionName: "foo",
		},
	}

	triggerClientset := kubelessFake.NewSimpleClientset(&f, &cjtrigger)

	cronjob := batchv1beta1.CronJob{
		ObjectMeta: myNsFoo,
	}
	clientset := fake.NewSimpleClientset(&cronjob)

	controller := CronJobTriggerController{
		clientset:      clientset,
		kubelessclient: triggerClientset,
		logger:         logrus.WithField("controller", "cronjob-trigger-controller"),
	}

	// no-op for when the function is not deleted
	err := controller.functionAddedDeletedUpdated(&f, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	list, err := controller.kubelessclient.KubelessV1beta1().CronJobTriggers("myns").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("Trigger should be deleted from list: %v", list.Items)
	}
}

func TestCronJobTriggerObjChanged(t *testing.T) {
	type testObj struct {
		old             *kubelessApi.CronJobTrigger
		new             *kubelessApi.CronJobTrigger
		expectedChanged bool
	}
	t1 := metav1.Time{
		Time: time.Now(),
	}
	t2 := metav1.Time{
		Time: time.Now(),
	}
	testObjs := []testObj{
		{
			old:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			new:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			expectedChanged: false,
		},
		{
			old:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &t1}},
			new:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &t2}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}},
			new:             &kubelessApi.CronJobTrigger{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.CronJobTrigger{Spec: kubelessApi.CronJobTriggerSpec{Schedule: "* * * * *"}},
			new:             &kubelessApi.CronJobTrigger{Spec: kubelessApi.CronJobTriggerSpec{Schedule: "* * * * *"}},
			expectedChanged: false,
		},
		{
			old:             &kubelessApi.CronJobTrigger{Spec: kubelessApi.CronJobTriggerSpec{Schedule: "*/10 * * * *"}},
			new:             &kubelessApi.CronJobTrigger{Spec: kubelessApi.CronJobTriggerSpec{Schedule: "* * * * *"}},
			expectedChanged: true,
		},
	}
	for _, to := range testObjs {
		changed := cronJobTriggerObjChanged(to.old, to.new)
		if changed != to.expectedChanged {
			t.Errorf("%v != %v expected to be %v", to.old, to.new, to.expectedChanged)
		}
	}
}
