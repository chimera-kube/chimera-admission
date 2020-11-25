# chimera-admission

`chimera-admission` is a
[Kubernetes dynamic admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
that loads Chimera Policies to validate admission requests.

Chimera Policies are simple [WebAssembly](https://webassembly.org/)
modules.

> **Note well:** the Chimera Project is in its early days. Many
> things are going to change. It's not meant to be used in production.
>
> We also plan to create a Kubernetes controller to simplify the management
> of Chimera Policies.

## Running chimera-admission

You can either build `chimera-admission` from sources (see the dedicated section
near the bottom of the file) or you can use the container image we maintain
inside of our [GitHub Container Registry](https://github.com/orgs/chimera-kube/packages/container/package/chimera-admission).

The [`deployment`](https://github.com/chimera-kube/chimera-admission/tree/main/deployment)
directory inside of this repository shows how to deploy `chimera-admission`
on top of a Kubernetes cluster.

## Configuring chimera-admission

The behaviour of `chimera-admission` can be tuned either via cli flags or
environment variables. All the environment variables have the `CHIMERA_` prefix.

### Referencing the Chimera Policy to be used

> **Note well:** at this time `chimera-admission` loads only one Chimera Policy.
> We have plans to allow users to dynamically set up Chimera Policies
> that `chimera-admission` will discover and load as needed.

The WASM file providing the Chimera Policy can be either loaded from
the local filesystem or it can be fetched from a remote location. The behaviour
depends on the URL format provided by the user:

* `file:///some/local/program.wasm`: load the policy from the local filesystem
* `https://some-host.com/some/remote/program.wasm`: download the policy from the
  remote http(s) server
* `registry://localhost:5000/project/artifact:some-version` download the policy
  from a OCI registry. The policy must have been pushed as an OCI artifact

### Policy tuning

Chimera Policies can be configured via environment variables. The `chimera-admission`
controller takes care of forwarding the environment variables from the host
system to the WASM runtime.

The controller automatically forwards all the environment variables that
have the `CHIMERA_EXPORT_` prefix. These environment variables are forwarded
with the `CHIMERA_EXPORT_` prefix stripped. For example, the `CHIMERA_EXPORT_ALLOWED_GROUPS`
will be forwarded to the Chimera Policy as `ALLOWED_GROUPS`.

### Kubernetes registration

The `chimera-admission` controller automatically takes care of registering
itself against the Kubernetes API server.

Kubernetes requires all the dynamic admission controllers to be secured with a
TLS certificate. The CA bundle used to generate the certificate used by the
controller must be provided to Kubernetes when the controller is registered.

The user can provide the TLS certificates and the CA bundle to use. When nothing
is specified, `chimera-admission` will automatically generate a CA and use it
to sign a TLS certificate.

> **Note well:** right now `chimera-admission` doesn't implement certificate rotation.

The `chimera-admission` will automatically register itself using the name provided
by the user, or use a self-generated one.

> **Note well:** the `ValidatingWebhookConfiguration` object is not deleted when
> the `chimera-admission` is terminated.

## Example

You need a Kubernetes cluster running and accessible through a `kubeconfig` file.
This can be done quickly using k3s.

The following commands download k3s and then run it locally:

```shell
$ wget https://github.com/rancher/k3s/releases/download/v1.19.4%2Bk3s1/k3s
$ chmod +x k3s
$ ./k3s server --disable-agent
```

Now we can start a `chimera-admission` instance that uses
[this Chimera Policy](https://github.com/chimera-kube/pod-toleration-policy)
to validate Pod operations.

The policy will be downloaded from
[here](https://github.com/orgs/chimera-kube/packages/container/package/policies%2Fpod-toleration).

```shell
$ CHIMERA_RESOURCES=pods \
  CHIMERA_EXPORT_TAINT_KEY=dedicated \
  CHIMERA_EXPORT_TAINT_VALUE=tenantA \
  CHIMERA_EXPORT_ALLOWED_GROUPS=system:masters \
  CHIMERA_WASM_URI=registry://ghcr.io/chimera-kube/policies/pod-toleration:v0.0.2 \
  KUBECONFIG=$HOME/.kube/k3s.yaml \
  ./chimera-admission-amd64
```

Now we can see the policy in action by creating the following Pod:

```shell
$ k3s kubectl apply -f - <<EOF
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
  - key: "dedicated"
    operator: "Equal"
    value: "tenantA"
    effect: "NoSchedule"
EOF
```

The `chimera-admission` instance will accept the creation request because the
`kubeconfig` used by k3s authenticates us as user named `kubernetes-admin` who
belongs to the `sytem:masters` and to the `system:authenticated` groups.

Let's remove the Pod now, so that we can make one last test:

```shell
$ k3s kubectl delete pod nginx
```

Stop the previous admission server execution, and re-run it with
a different tuning of the Chimera Policy:

```shell
$ CHIMERA_RESOURCES=pods \
  CHIMERA_EXPORT_TAINT_KEY=dedicated \
  CHIMERA_EXPORT_TAINT_VALUE=tenantA \
  CHIMERA_EXPORT_ALLOWED_GROUPS=trusted-users \
  CHIMERA_WASM_URI=registry://ghcr.io/chimera-kube/policies/pod-toleration:v0.0.2 \
  KUBECONFIG=$HOME/.kube/k3s.yaml \
  ./chimera-admission-amd64
```

Now the policy accepts this toleration only when a user who belongs to the
`trusted-users` group is the author of the request.

Let's create the same Pod one last time:

```shell
$ k3s kubectl apply -f - <<EOF
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

The admission controller is properly working: the creation request has
been rejected because it's not done by a user who belongs to one of the
valid groups.

## Building

The `chimera-admission` binary can be built in this way:

```shell
$ # Build x86_64 binary
$ make chimera-admission-amd64
```
