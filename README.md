# Web-scale Data Management Project

## Design


## API


## Environment Variables
- `ORDER_SERVICE_URL`, `STOCK_SERVICE_URL`, `PAYMENT_SERVICE_URL`: `http://host:port/`  
  The leading protocol and ending slash are necessary.

- `REDIS_ADDRS`: `host:port` or `host:port,host:port,...,host:port`  
  Comma separated list of strings to sharded redis. DO NOT add any space or ending comma.

- `REDIS_PASSWORD`: `pwd`  
  For simplicity, the passwords are required to be the same.

- `REDIS_DB`: `0`  
  For simplicity, the dbs are required to be the same.

- `MACHINE_ID`: `1/3`  
  Index of machine (ie k8s deployment of the same image) slash number of machines. Index starts from 1. Note that the default routing is a simple identical mapping from machine id to redis shard id, thus requiring the number of machines identical to that of redis databases. If they disagree, you should implement your own hashing algorithm.

- `WDM_DEBUG`: `1` or `0`  



## Test
Use docker to deploy the project locally and run
```bash
cd test
pip install -r requirements.txt
python test_microservices.py
```
to do the correctness test.

To perform consistency test and stress test, see [wdm-project-benchmark](https://github.com/wdm23-5/wdm-project-benchmark).



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
helm delete redis-order
helm delete redis-stock
helm delete redis-payment
helm install redis-order bitnami/redis -f k8s/helm/redis-helm-values.yaml
helm install redis-stock bitnami/redis -f k8s/helm/redis-helm-values.yaml
helm install redis-payment bitnami/redis -f k8s/helm/redis-helm-values.yaml

# deploy
kubectl delete -f k8s/kube/.
kubectl apply -f k8s/kube/.

# test
minikube tunnel
```

The REST APIs are available directly under `localhost`.


### Google Kubernetes Engine
Follow the [quickstart](https://cloud.google.com/kubernetes-engine/docs/deploy-app-cluster) up to the section "Create a GKE cluster". Then in the shell execute
```bash
prefix="https://raw.githubusercontent.com/wdm23-5/wdm-project-go/main/"

# helm chart
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx

helm install redis-order bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml
helm install redis-stock bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml
helm install redis-payment bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml
# wait a moment for the databases to be ready

# deploy
kubectl apply -f ${prefix}gke/kube/order-app.yaml
kubectl apply -f ${prefix}gke/kube/stock-app.yaml
kubectl apply -f ${prefix}gke/kube/payment-app.yaml
# wait a moment for the services to be ready
kubectl apply -f ${prefix}gke/kube/ingress-service.yaml
# wait a moment for the ingress to be ready

# view external ip
echo $(kubectl get ingress ingress-service -ojson | jq -r '.status.loadBalancer.ingress[].ip')
```

If you want to deploy the project at scale, check out the scripts under `gke-scale`.
