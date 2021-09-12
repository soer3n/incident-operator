# exemplary situation

There are two worker nodes with respectively one daemonset and deployment deployed in a kubernetes cluster.

```

$ kubectl get nodes 
NAME                              STATUS   ROLES                  AGE     VERSION
mngt-control-plane-1              Ready    control-plane,master   2d      v1.20.5
mngt-mngt-pool-7c46bb775f-fghln   Ready    <none>                 2d      v1.20.5
mngt-mngt-pool-7c46bb775f-jpxv9   Ready    <none>                 2d      v1.20.5

$ kubectl get deployments.apps
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
grafana   1/1     1            1           2d

$ kubectl get daemonsets.apps
NAME                  DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
loki-stack-promtail   2         2         2       2            2           <none>          2d

$ kubectl get pods -o wide
NAME                        READY   STATUS    RESTARTS   AGE   IP            NODE                              NOMINATED NODE   READINESS GATES
grafana-5c7b49968d-nftz2    1/1     Running   0          2d    xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>           <none>
loki-stack-0                1/1     Running   0          2d    xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-jpxv9   <none>           <none>
loki-stack-promtail-b2l26   1/1     Running   0          2d    xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-jpxv9   <none>           <none>
loki-stack-promtail-wrkjx   1/1     Running   0          2d    xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>           <none>

```

## start the quarantine

For a detailed explanation of configurable fields go to [components](docs/COMPONENTS.md). After applying quarantine resource file from [examples](examples/basic.yaml) one worker node will be cordoned, the pod of the daemonset on the configured nodes will be relabeled to isolate, a debug pod will be deployed, a toleration will be added to the daemonset, a taint will be added to cordoned node and at the end the node will be drained. 

```

$ kubectl apply -f examples/basic.yaml 
quarantine.ops.soer3n.info/quarantine-sample created

$ kubectl get pods -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP            NODE                              NOMINATED NODE   READINESS GATES
grafana-5c7b49968d-2q6ll    0/1     Pending   0          60s     <none>        <none>                            <none>           <none>
loki-stack-0                1/1     Running   0          2d      xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-jpxv9   <none>           <none>
loki-stack-promtail-9sfzl   1/1     Running   0          57s     xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>           <none>
loki-stack-promtail-dg4sf   1/1     Running   0          2m12s   xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>           <none>
loki-stack-promtail-hct2x   1/1     Running   0          36s     xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-jpxv9   <none>           <none>

$ kubectl get pod -A -l quarantine=true -o wide
NAMESPACE     NAME                        READY   STATUS    RESTARTS   AGE     IP            NODE                              NOMINATED NODE   READINESS GATES
kube-system   quarantine-debug            1/1     Running   0          2m44s   xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>    <none>
default       loki-stack-promtail-dg4sf   1/1     Running   0          3m55s   xx.xx.xx.xx   mngt-mngt-pool-7c46bb775f-fghln   <none>    <none>

$ kubectl get nodes -o wide
NAME                              STATUS                     ROLES                  AGE    VERSION   INTERNAL-IP   EXTERNAL-IP  ...
mngt-control-plane-1              Ready                      control-plane,master   2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx  
mngt-mngt-pool-7c46bb775f-fghln   Ready,SchedulingDisabled   <none>                 2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx 
mngt-mngt-pool-7c46bb775f-jpxv9   Ready                      <none>                 2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx   

```

## debugging while node is drained

Now you can open a terminal in the debung pod. Host filesystem is mounted into this container and several network tools are available. You can also use your own image.

```

$ kubectl exec -ti -n kube-system quarantine-debug -- bash
bash-5.1# 

```


## stop the quarantine

If work is finished you can reschedule workloads, remove isolated and debug pods by deleting the quarantine resource.

```

$ kubectl delete -f examples/basic.yaml 
quarantine.ops.soer3n.info "quarantine-sample" deleted

$ kubectl get nodes -o wide
NAME                              STATUS   ROLES                  AGE    VERSION   INTERNAL-IP   EXTERNAL-IP  ...    
mngt-control-plane-1              Ready    control-plane,master   2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx     
mngt-mngt-pool-7c46bb775f-fghln   Ready    <none>                 2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx
mngt-mngt-pool-7c46bb775f-jpxv9   Ready    <none>                 2d     v1.20.5   xx.xx.xx.xx   xx.xx.xx.xx

$ kubectl get pod -A -l quarantine=true -o wide
No resources found

$ kubectl get pod -n loki-stack -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP               NODE                              NOMINATED NODE   READINESS GATES
grafana-5c7b49968d-2q6ll    1/1     Running   0          12m     xx.xx.xx.xx     mngt-mngt-pool-7c46bb775f-fghln    <none>           <none>
loki-stack-0                1/1     Running   0          2d      xx.xx.xx.xx     mngt-mngt-pool-7c46bb775f-jpxv9    <none>           <none>
loki-stack-promtail-hrmw8   1/1     Running   0          2m7s    xx.xx.xx.xx     mngt-mngt-pool-7c46bb775f-fghln    <none>           <none>
loki-stack-promtail-r9s5x   1/1     Running   0          2m22s   xx.xx.xx.xx     mngt-mngt-pool-7c46bb775f-jpxv9    <none>           <none>

```
