module github.com/chimera-kube/chimera-admission

go 1.15

require (
	github.com/bytecodealliance/wasmtime-go v0.21.0
	github.com/chimera-kube/chimera v0.0.0-00010101000000-000000000000
	github.com/engineerd/wasm-to-oci v0.1.1
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.3.0
	k8s.io/api v0.18.6
)

replace github.com/chimera-kube/chimera => ../chimera
