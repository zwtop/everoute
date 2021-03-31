module github.com/smartxworks/lynx

go 1.15

require (
	github.com/agiledragon/gomonkey v2.0.2+incompatible
	github.com/contiv/libovsdb v0.0.0
	github.com/contiv/ofnet v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/contiv/libOpenflow => github.com/echkenluo/libOpenflow v0.0.0-20210303025312-9765e623c87e
	github.com/contiv/libovsdb => github.com/smartxworks/libovsdb v0.0.0-20210326110222-6c508538aa65
	github.com/contiv/ofnet => github.com/smartxworks/ofnet v0.0.0-20210331075237-63518757ed1a
	github.com/osrg/gobgp => github.com/zwtop/gobgp v0.0.0-20210127101833-12edfc1f4514
)
