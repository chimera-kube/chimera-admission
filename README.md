# chimera-admission

`chimera-admission` is a program that allows you to register and run
Kubernetes admission webhooks that will load a WASM environment to
perform admission control dynamically.

## Requirements

For testing `chimera-admission` you need a Kubernetes cluster running
and accessible through a `kubeconfig` file.

You can start a k3s instance by downloading k3s and executing the k3s
server in your machine, like so:

```shell
~ $ wget https://github.com/rancher/k3s/releases/download/v1.19.4%2Bk3s1/k3s
~ $ chmod +x k3s
~ $ ./k3s server --disable-agent
```

## Building chimera-admission

Build the `chimera-admission-amd64` binary:

```shell
$ make chimera-admission-amd64
```

## Running chimera-admission

Run the `chimera-admission` server, that will load the provided WASM
module, and register the admission webhook on the Kubernetes API
automatically.

```shell
$ CHIMERA_RESOURCES=pods \
  CHIMERA_EXPORT_TOLERATION_KEY=example-key \
  CHIMERA_EXPORT_TOLERATION_OPERATOR=Exists \
  CHIMERA_EXPORT_TOLERATION_EFFECT=NoSchedule \
  CHIMERA_EXPORT_ALLOWED_GROUPS=system:authenticated \
  CHIMERA_WASM_URI=file://$PWD/wasm-examples/pod-toleration-policy/pod-toleration-policy.wasm \
  KUBECONFIG=$HOME/.kube/k3s.yaml \
  ./chimera-admission-amd64
```

Currently, WASM modules can be provided to the `CHIMERA_WASM_URI`
environment variable in three different ways:

* Local filesystem: `CHIMERA_WASM_URI=file:///some/local/program.wasm`
* HTTP(s): `CHIMERA_WASM_URI=https://some-host.com/some/remote/program.wasm`
* OCI Registry: `CHIMERA_WASM_URI=registry://localhost:5000/project/artifact:some-version`

At this time the `chimera-admission` project only loads one WASM
program, but we have plans to allow users to dynamically set up WASM
modules that `chimera-admission` will discover and load as needed.

## Trying the loaded WASM module

After you have started the `chimera-admission` server on the previous
step, you will see that creating the following pod will be allowed:

```shell
~ $ k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  tolerations:
  - key: "example-key"
    operator: "Exists"
    effect: "NoSchedule"
EOF
```

Remove the pod after it has been created, so we can try to see how
it's rejected afterwards. Run:

```shell
~ $ k3s kubectl delete pod nginx
```

Stop the previous admission server execution, and re-run it with
exposed environment variables slightly changed, like so:

```shell
$ CHIMERA_RESOURCES=pods \
  CHIMERA_EXPORT_TOLERATION_KEY=example-key \
  CHIMERA_EXPORT_TOLERATION_OPERATOR=Exists \
  CHIMERA_EXPORT_TOLERATION_EFFECT=NoSchedule \
  CHIMERA_EXPORT_ALLOWED_GROUPS=some-other-group \
  CHIMERA_WASM_URI=file://$PWD/wasm-examples/pod-toleration-policy/pod-toleration-policy.wasm \
  KUBECONFIG=$HOME/.kube/k3s.yaml \
  ./chimera-admission-amd64
```

Now, if you try to create the pod again you will get the following error:

```shell
~ $ k3s kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  tolerations:
  - key: "example-key"
    operator: "Exists"
    effect: "NoSchedule"
EOF
Error from server: error when creating "STDIN": admission webhook "rule-0.wasm.admission.rule" denied the request: User not allowed to create Pod objects with toleration: key: example-key, operator: Exists, effect: NoSchedule)
```

This is the sample WASM program `pod-toleration-policy.wasm` rejecting
the pod creation request, because the group in the creation request is
not included in the allowed groups provided in the
server `CHIMERA_EXPORT_ALLOWED_GROUPS` environment variable.
