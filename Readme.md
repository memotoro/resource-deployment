# SeldonIO Resource Deployment

This exercise assumes that minikube, kubernetes, docker, helm and GoLang are already installed.

* Minikube v1.7.3 on Ubuntu 16.04
* Docker Engine v19.03.6
* Kubernetes v1.17.3
* Helm v3.1.1
* GoLang go1.13.6

## Get the environment ready

First of all, start minikube server with the following command. The flag `--apiserver-port` will set up the api port with the given value. That value is required for later.

```
minikube start --apiserver-port=8443
```

Check the status of the minikube cluster once the previous step has finished.

```
minikube status
```

The output should be similar to the following:

```
host: Running
kubelet: Running
apiserver: Running
```

Get the minikube cluster IP address with the following command. The output value is required for later.

```
minikube ip
```

Tht output should be similar to the following:

```
192.168.99.100
```

Install `seldon-core` with the following `helm` command. Have a look to the `namespace` that is provided. We need that `namespace` value for later.

```
kubectl create namespace seldon-system
```

```
helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --namespace seldon-system
```

Once everything is installed, check that everything is OK. The `pods` should be running and `deployments/replicasets` should be available. Run the following command and provide the right `namespace`

```
kubectl get all --namespace seldon-system
```

For simplicity, I'll reuse the `serviceaccount` and `token` that the `seldon-core` is using. The same could be done by creating a new `serviceaccount`, with a `clusterrolebinding` and `clusterrole` with the right verbs we need for creating, deleting and getting status from a resource, but again for simplicity I'll reuse the `serviceaccount`.

First of all, get the `clusterrolebinding` in the cluster with the following command:

```
kubectl get clusterrolebinding
```

The output should be similar to the following:

```
NAME                                                   AGE
cluster-admin                                          6m17s
kubeadm:kubelet-bootstrap                              6m15s
kubeadm:node-autoapprove-bootstrap                     6m15s
kubeadm:node-autoapprove-certificate-rotation          6m15s
kubeadm:node-proxier                                   6m15s
minikube-rbac                                          6m14s
seldon-manager-rolebinding-seldon-system               35s
seldon-manager-sas-rolebinding-seldon-system           35s
....
```

In the example above, we are interested in the `clusterrolebinding` called `seldon-manager-rolebinding-seldon-system`. Describe the `culsterrolebinding` with the following command:

```
kubectl describe clusterrolebinding seldon-manager-rolebinding-seldon-system
```

The output should be similar to the following:

```
Name:         seldon-manager-rolebinding-seldon-system
Labels:       app=seldon
              app.kubernetes.io/instance=seldon-core
              app.kubernetes.io/name=seldon-core-operator
              app.kubernetes.io/version=1.0.2
Annotations:  <none>
Role:
  Kind:  ClusterRole
  Name:  seldon-manager-role-seldon-system
Subjects:
  Kind            Name            Namespace
  ----            ----            ---------
  ServiceAccount  seldon-manager  seldon-system
```

In the output above, we are interested in the `serviceaccount` called `seldon-manager` and the role called `seldon-manager-role-seldon-system`

Let's explore the role first with the following command

```
kubectl describe clusterrole seldon-manager-role-seldon-system
```

The output should be similar to the following:

```
Name:         seldon-manager-role-seldon-system
Labels:       app=seldon
              app.kubernetes.io/instance=seldon-core
              app.kubernetes.io/name=seldon-core-operator
              app.kubernetes.io/version=1.0.2
Annotations:  <none>
PolicyRule:
  Resources                                               Non-Resource URLs  Resource Names  Verbs
  ---------                                               -----------------  --------------  -----
  services                                                []                 []              [create delete get list patch update watch]
  deployments.apps                                        []                 []              [create delete get list patch update watch]
  horizontalpodautoscalers.autoscaling                    []                 []              [create delete get list patch update watch]
  seldondeployments.machinelearning.seldon.io             []                 []              [create delete get list patch update watch]
  destinationrules.networking.istio.io                    []                 []              [create delete get list patch update watch]
  virtualservices.networking.istio.io                     []                 []              [create delete get list patch update watch]
  services.v1                                             []                 []              [create delete get list patch update watch]
  namespaces                                              []                 []              [get list watch]
  namespaces.v1                                           []                 []              [get list watch]
  deployments.apps/status                                 []                 []              [get patch update]
  horizontalpodautoscalers.autoscaling/status             []                 []              [get patch update]
  seldondeployments.machinelearning.seldon.io/finalizers  []                 []              [get patch update]
  seldondeployments.machinelearning.seldon.io/status      []                 []              [get patch update]
  destinationrules.networking.istio.io/status             []                 []              [get patch update]
  virtualservices.networking.istio.io/status              []                 []              [get patch update]
  services.v1/status                                      []                 []              [get patch update]
```

From the output above we need to pay attention to the verbs allowed to the resource `seldondeployments.machinelearning.seldon.io` as this is the kind of resource we are going to deploy with the API and this Go program.

Describe the `serviceaccount` called `seldon-manager` with the following command:

```
kubectl describe serviceaccount seldon-manager --namespace seldon-system
```

The output should be similar to the following:

```
Name:                seldon-manager
Namespace:           seldon-system
Labels:              app=seldon
                     app.kubernetes.io/instance=seldon-core
                     app.kubernetes.io/name=seldon-core-operator
                     app.kubernetes.io/version=1.0.2
Annotations:         <none>
Image pull secrets:  <none>
Mountable secrets:   seldon-manager-token-brfgg
Tokens:              seldon-manager-token-brfgg
Events:              <none>
```

In the example above, the name of the token is `seldon-manager-token-brfgg`. Get the token value for the `serviceaccount` with the following command:

```
kubectl describe secret seldon-manager-token-brfgg --namespace seldon-system
```

The output should be similar to the following:

```
Name:         seldon-manager-token-brfgg
Namespace:    seldon-system
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: seldon-manager
              kubernetes.io/service-account.uid: b2ea61f5-341d-4aeb-9f9c-3474866404d9

Type:  kubernetes.io/service-account-token

Data
====
ca.crt:     1066 bytes
namespace:  13 bytes
token:      eyJhbGciOiJSUzI1NiIsImtpZCI6IkhCUi1BREtmUWFjM3d2NEtLQ3NuRnFydy1RZk1ERXRRS0FCSE13bjQ0ZWcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJzZWxkb24tc3lzdGVtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InNlbGRvbi1tYW5hZ2VyLXRva2VuLWJyZmdnIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6InNlbGRvbi1tYW5hZ2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiYjJlYTYxZjUtMzQxZC00YWViLTlmOWMtMzQ3NDg2NjQwNGQ5Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OnNlbGRvbi1zeXN0ZW06c2VsZG9uLW1hbmFnZXIifQ.XqUpfGXAimC5KJ-D9PjpK_YHQsWhI-U4m-Vbsmj3-CHxJh6QC4siWMZnR_odQhwqKTejgx8NhMYz7dBaAsW1G0r8x2YR_819YddQgI3-QIwg_d-n1jC6j2RxarNYf6KNfjBKwiMbnf5j_KRDfe6jQwYlqAsHnn-iNcXeCyFwf1yCt6NPPxH93-sjRxyNAsyIiuwGX1OlMb3eCsqE3XY0keWnL1gUn2u8s_SvdoAwi1-oKKxASb2AlCbpSA4pq26YuSfMApSuDi1BhU9JoHi9zkrrYfpJz_YbgbDdQfPjr1lU9d1zU7ttBpxFW-G6p9mYnzbH8l2PpjRuaSWD3ZEcww
```

Get the entire value for the attribute `token` as it is required for later.

## Get the resource file

A sample seldon resource is located in `data/seldon-model.json`. It was extracted from [here](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/notebooks/resources/model.json) . However, the resource file could be provided either with a file in the local computer or a remote file available through the Internet.

## Running the Go command line

Download the project and then execute in the root folder the following:

```
go run main.go --help
```

That will display all the different options that you can provide to the tool to do the job. There are some default values you can leave or override if required. The values should be replaced with the right value you were getting along with this tutorial in all the previous commands.

```
usage: main [<flags>]

Flags:
  --help                       Show context-sensitive help (also try --help-long and --help-man).
  --host-api="192.168.99.100"  Hostname or IP address for the kubernetes API
  --port-api="8443"            Port number for the kubernetes API
  --api-timeout=5s             Timeout in time.Duration for http calls to Kubernetes API
  --api-token=""               API kubernetes token with the right permissions for the API
  --resource-file=""           Location of file with the resource to be created. Could be a file path or http url
  --waiting-time=5s            Time in seconds to wait between API call
  --max-attempts=20            Maximum number of re-tries before terminating the task
  --namespace="seldon-system"  Namespace for kubernetes
  --verbose                    Verbose logging

```


Running with a local file

```
go run main.go --host-api="192.168.99.100" \
--port-api="8443" \
--namespace="seldon-system" \
--resource-file="./data/seldon-model.json" \
--api-token="eyJhbGciOiJSUzI1NiIsImtpZCI6IkhCUi1BREtmUWFjM3d2NEtLQ3NuRnFydy1RZk1ERXRRS0FCSE13bjQ0ZWcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJzZWxkb24tc3lzdGVtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InNlbGRvbi1tYW5hZ2VyLXRva2VuLWJyZmdnIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6InNlbGRvbi1tYW5hZ2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiYjJlYTYxZjUtMzQxZC00YWViLTlmOWMtMzQ3NDg2NjQwNGQ5Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OnNlbGRvbi1zeXN0ZW06c2VsZG9uLW1hbmFnZXIifQ.XqUpfGXAimC5KJ-D9PjpK_YHQsWhI-U4m-Vbsmj3-CHxJh6QC4siWMZnR_odQhwqKTejgx8NhMYz7dBaAsW1G0r8x2YR_819YddQgI3-QIwg_d-n1jC6j2RxarNYf6KNfjBKwiMbnf5j_KRDfe6jQwYlqAsHnn-iNcXeCyFwf1yCt6NPPxH93-sjRxyNAsyIiuwGX1OlMb3eCsqE3XY0keWnL1gUn2u8s_SvdoAwi1-oKKxASb2AlCbpSA4pq26YuSfMApSuDi1BhU9JoHi9zkrrYfpJz_YbgbDdQfPjr1lU9d1zU7ttBpxFW-G6p9mYnzbH8l2PpjRuaSWD3ZEcww"
```

Running with a remote file

```
go run main.go --host-api="192.168.99.100" \
--port-api="8443" \
--namespace="seldon-system" \
--resource-file="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/notebooks/resources/model.json" \
--api-token="eyJhbGciOiJSUzI1NiIsImtpZCI6IkhCUi1BREtmUWFjM3d2NEtLQ3NuRnFydy1RZk1ERXRRS0FCSE13bjQ0ZWcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJzZWxkb24tc3lzdGVtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InNlbGRvbi1tYW5hZ2VyLXRva2VuLWJyZmdnIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6InNlbGRvbi1tYW5hZ2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiYjJlYTYxZjUtMzQxZC00YWViLTlmOWMtMzQ3NDg2NjQwNGQ5Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OnNlbGRvbi1zeXN0ZW06c2VsZG9uLW1hbmFnZXIifQ.XqUpfGXAimC5KJ-D9PjpK_YHQsWhI-U4m-Vbsmj3-CHxJh6QC4siWMZnR_odQhwqKTejgx8NhMYz7dBaAsW1G0r8x2YR_819YddQgI3-QIwg_d-n1jC6j2RxarNYf6KNfjBKwiMbnf5j_KRDfe6jQwYlqAsHnn-iNcXeCyFwf1yCt6NPPxH93-sjRxyNAsyIiuwGX1OlMb3eCsqE3XY0keWnL1gUn2u8s_SvdoAwi1-oKKxASb2AlCbpSA4pq26YuSfMApSuDi1BhU9JoHi9zkrrYfpJz_YbgbDdQfPjr1lU9d1zU7ttBpxFW-G6p9mYnzbH8l2PpjRuaSWD3ZEcww"
```

The output should be similar to the following:

```
2020/02/29 18:55:34 Resource Name [SeldonDeployment] - Resource Kind [seldon-model] created.
2020/02/29 18:55:39 Resource Name [seldon-model] - Resource State [Creating]. Waiting 5s before attempt 1 out of 20
2020/02/29 18:55:44 Resource Name [seldon-model] - Resource State [Creating]. Waiting 5s before attempt 2 out of 20
2020/02/29 18:55:49 Resource Name [seldon-model] - Resource State [Creating]. Waiting 5s before attempt 3 out of 20
2020/02/29 18:55:54 Resource Name [seldon-model] - Resource State [Creating]. Waiting 5s before attempt 4 out of 20
2020/02/29 18:55:59 Resource Name [seldon-model] - Resource State [Creating]. Waiting 5s before attempt 5 out of 20
2020/02/29 18:56:04 Resource Name [seldon-model] - Resource State [Available]. Waiting 5s before deleting resource
2020/02/29 18:56:09 Resource Name [seldon-model] was deleted
```
