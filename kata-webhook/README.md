# Kata Admission controller webhook

Implement a simple admission controller webhook to annotate pods with the 
kata runtime class.

## How to use the admission controller

```
docker build -t mcastelino/kubewebhook-pod-annotate-example:latest
./create_certs.sh
```

Note: The docker image location can be changed and the image needs to be
published if the webhook needs to work.

On a single machine cluster change the `imagePullPolicy` to use the locally 
built image.


## Making Kata the default runtime

Today in `crio.conf` runc is the default runtime when a user does not specify
`runtimeClass` in the pod spec. If you want to run a cluster where kata is used
by default, except for workloads we know for sure will not work with kata, use
the [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks)
and sample admission controller we created by running -

`kubectl apply -f deploy/`

The [admission webhook](deploy/webhook-registration.yaml)
is setup to exclude certian namespaces from being run with Kata using filters on namespace labels.

```yaml
    namespaceSelector:
      matchExpressions:
        -  {key: "kata", operator: NotIn, values: ["false"]}
```

The rook operators for example are marked as such

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: rook-ceph-system
  labels:
    kata: "false"
```

Pods not explicitly excluded by the namespace filter are dynamically tagged to
run with Kata with some [exceptions](https://github.com/mcastelino/kata-webhook/main.go#L25) -

* `hostNetwork: true`
* `rook-ceph` and `rook-ceph-system` namespaces (buggy)

Other pod properties will be added as exceptions in future.

