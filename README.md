# Web-scale Data Management Project


## Deployment
### docker
Run `docker-compose up --build` in the base folder. The REST APIs are available at `localhost:8000`.

### minikube
```bash
# minikube
minikube delete
minikube start --memory=16384 --cpus=8
minikube addons enable ingress

# helm chart
./deploy-charts-minicube.sh

# deploy
kubectl delete -f k8s/.
kubectl apply -f k8s/.

# test
minikube tunnel
```

The REST APIs are available directly under `localhost`.
