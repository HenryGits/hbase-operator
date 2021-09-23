module github.com/HenryGits/hbase-operator

go 1.16

require (
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/controller-runtime v0.9.2
	gitee.com/dmcca/gotools v0.0.2-0.20210817112322-50db2523d334
)

replace (
	gitee.com/dmcca/compass-tenant v0.0.0-20210824123435-eb963b6337b6 => C:\Apps\go\gopath\compass-tenant@v0.0.0
	gitee.com/dmcca/gotools v0.0.2-0.20210817112322-50db2523d334 => C:\Apps\go\gopath\gotools@v0.0.0
)
