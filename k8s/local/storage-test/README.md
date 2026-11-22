# Temporary PVC persistence exercise

This exercise uses a disposable 16 MiB PVC in namespace `selena-dev` and the small `busybox:1.36.1` image. It does not mount or access the Docker PostgreSQL volume.

Apply the claim and writer Pod, then wait for the writer to complete:

```sh
kubectl --context=selena-local apply -f k8s/local/storage-test/pvc.yaml
kubectl --context=selena-local wait --for=jsonpath='{.status.phase}'=Bound pvc/users-pvc-test-data -n selena-dev --timeout=90s
kubectl --context=selena-local apply -f k8s/local/storage-test/pod-writer.yaml
kubectl --context=selena-local wait --for=jsonpath='{.status.phase}'=Succeeded pod/users-pvc-writer -n selena-dev --timeout=120s
kubectl --context=selena-local logs pod/users-pvc-writer -n selena-dev
```

Delete only the writer Pod, create the reader Pod against the same PVC, and verify its log:

```sh
kubectl --context=selena-local delete -f k8s/local/storage-test/pod-writer.yaml
kubectl --context=selena-local apply -f k8s/local/storage-test/pod-reader.yaml
kubectl --context=selena-local wait --for=jsonpath='{.status.phase}'=Succeeded pod/users-pvc-reader -n selena-dev --timeout=120s
kubectl --context=selena-local logs pod/users-pvc-reader -n selena-dev
```

After the marker is confirmed, remove only the diagnostic Pods and PVC:

```sh
kubectl --context=selena-local delete -f k8s/local/storage-test/pod-reader.yaml --ignore-not-found
kubectl --context=selena-local delete -f k8s/local/storage-test/pvc.yaml
```

The default Minikube StorageClass provisions the volume. Do not use `hostPath`, change the StorageClass, or run Docker cleanup commands for this exercise.
