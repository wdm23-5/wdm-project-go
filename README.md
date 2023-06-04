# Web-scale Data Management Project

## API


## Test
Use docker to deploy the project locally and run
```bash
cd test
pip install -r requirements.txt
python test_microservices.py
```


## Deployment
You can do either of the followings to deploy the project.

### docker
Run `docker-compose up --build` in the base folder. The REST APIs are available at `localhost:8000`.

### minikube
```bash
# minikube
minikube delete
minikube start --memory=16384 --cpus=8
minikube addons enable ingress

# helm chart
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm delete <name of previous redis>
helm install redis-order bitnami/redis -f helm-config/redis-helm-values.yaml
helm install redis-stock bitnami/redis -f helm-config/redis-helm-values.yaml
helm install redis-payment bitnami/redis -f helm-config/redis-helm-values.yaml

# deploy
kubectl delete -f k8s/.
kubectl apply -f k8s/.

# test
minikube tunnel
```

The REST APIs are available directly under `localhost`.
