# Deployment

To deploy, you'll need [Helm](https://helm.sh/docs/intro/install/) and [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube) (for convenience)

start your minikube
```
minikube start
```

Once started, run
```
helm install unity-msg-svc ./unity-msg-chart
```

Wait for the pods to be healthy with
```
kubectl get pods
```
**This might take a few seconds as rabbitmq is slow to boot**

Finally, once all pods healthy, you can access the app with
```
minikube service unity-msg-svc-unity-msg-chart
```

# Useful commands
Accessing the redis-cli
```
kubectl exec -it unity-msg-svc-redis-master-0 -- redis-cli
```

Accessing the rabbitmq service
```
minikube service unity-msg-svc-rabbitmq
```

Seeing the logs of the server
```
kubectl get pods | awk '/unity-msg-svc-unity-msg-chart/{print $1}' | xargs kubectl logs -f
```