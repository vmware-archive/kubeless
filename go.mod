module github.com/kubeless/kubeless

go 1.12

require (
	github.com/Azure/go-autorest v8.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.16.26
	github.com/coreos/prometheus-operator v0.0.0-20171201110357-197eb012d973
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190130105114-cc9c99918988 // indirect
	github.com/gosuri/uitable v0.0.0-20160404203958-36ee7e946282
	github.com/imdario/mergo v0.3.7
	github.com/kubeless/cronjob-trigger v1.0.2
	github.com/kubeless/http-trigger v1.0.0
	github.com/kubeless/kafka-trigger v1.0.1
	github.com/kubeless/kinesis-trigger v0.0.0-20180817123215-a548c3d1cbd9
	github.com/kubeless/nats-trigger v0.0.0-20180817123246-372a5fa547dc
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/nats-io/gnatsd v1.4.1 // indirect
	github.com/nats-io/go-nats v1.7.0
	github.com/nats-io/nkeys v0.0.2 // indirect
	github.com/nats-io/nuid v1.0.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.4.0
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.1.1
	golang.org/x/build v0.0.0-20190111050920-041ab4dc3f9d // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.0.0-20180308224125-73d903622b73
	k8s.io/apiextensions-apiserver v0.0.0-20180327033742-750feebe2038
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
)
