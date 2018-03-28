package controller

import (
	"testing"
	"time"

	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessFake "github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"github.com/sirupsen/logrus"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHTTPFunctionAddedUpdated(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	f := kubelessApi.Function{
		ObjectMeta: myNsFoo,
	}

	httpTrigger := kubelessApi.HTTPTrigger{
		ObjectMeta: myNsFoo,
	}

	triggerClientset := kubelessFake.NewSimpleClientset(&f, &httpTrigger)

	ingress := extensionsv1beta1.Ingress{
		ObjectMeta: myNsFoo,
	}
	clientset := fake.NewSimpleClientset(&ingress)

	controller := HTTPTriggerController{
		clientset:      clientset,
		kubelessclient: triggerClientset,
		logger:         logrus.WithField("controller", "http-trigger-controller"),
	}

	// no-op for when the function is not deleted
	err := controller.functionAddedDeletedUpdated(&f, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	list, err := controller.kubelessclient.KubelessV1beta1().HTTPTriggers("myns").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].ObjectMeta.Name != "foo" {
		t.Errorf("Missing trigger in list: %v", list.Items)
	}
}

func TestHTTPFunctionDeleted(t *testing.T) {
	myNsFoo := metav1.ObjectMeta{
		Namespace: "myns",
		Name:      "foo",
	}

	f := kubelessApi.Function{
		ObjectMeta: myNsFoo,
	}

	httpTrigger := kubelessApi.HTTPTrigger{
		ObjectMeta: myNsFoo,
		Spec: kubelessApi.HTTPTriggerSpec{
			FunctionName: myNsFoo.Name,
		},
	}

	triggerClientset := kubelessFake.NewSimpleClientset(&f, &httpTrigger)

	ingress := extensionsv1beta1.Ingress{
		ObjectMeta: myNsFoo,
	}
	clientset := fake.NewSimpleClientset(&ingress)

	controller := HTTPTriggerController{
		clientset:      clientset,
		kubelessclient: triggerClientset,
		logger:         logrus.WithField("controller", "http-trigger-controller"),
	}

	err := controller.functionAddedDeletedUpdated(&f, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	list, err := controller.kubelessclient.KubelessV1beta1().HTTPTriggers("myns").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("Trigger should be deleted from list: %v", list.Items)
	}
}

func TestHTTPTriggerObjChanged(t *testing.T) {
	type testObj struct {
		old             *kubelessApi.HTTPTrigger
		new             *kubelessApi.HTTPTrigger
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
			old:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			new:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			expectedChanged: false,
		},
		{
			old:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &t1}},
			new:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &t2}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "1"}},
			new:             &kubelessApi.HTTPTrigger{ObjectMeta: metav1.ObjectMeta{ResourceVersion: "2"}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{HostName: "a"}},
			new:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{HostName: "a"}},
			expectedChanged: false,
		},
		{
			old:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{HostName: "a"}},
			new:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{HostName: "b"}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{TLSAcme: true}},
			new:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{TLSAcme: false}},
			expectedChanged: true,
		},
		{
			old:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{Path: "a"}},
			new:             &kubelessApi.HTTPTrigger{Spec: kubelessApi.HTTPTriggerSpec{Path: "b"}},
			expectedChanged: true,
		},
	}
	for _, to := range testObjs {
		changed := httpTriggerObjChanged(to.old, to.new)
		if changed != to.expectedChanged {
			t.Errorf("%v != %v expected to be %v", to.old, to.new, to.expectedChanged)
		}
	}
}
