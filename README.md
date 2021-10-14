### go-k8s-sample :

1. `go run main.go -kubeconfig <> -namespace <> -cluster_a_context <> -cluster_b_context <>`

    - `kubeconfig` - path to k8s config, if not given `/Users/<username>/.kube/config` will be set, optional.
    - `namespace` - namespace where apps are deployed in cluster A, required.
    - `cluster_a_context` - k8s context name of cluster A, if not given current-context will be taken, optional.
    - `cluster_b_context` - k8s context name of cluster B, required.

    **Note**: k8s config location in Mac : `/Users/<username>/.kube/config`, where cluster related context can be found.

2. For running test : `go test`
