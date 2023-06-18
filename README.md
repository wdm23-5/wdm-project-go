# Web-scale Data Management Project
This project implement the backend for an online e-commerce system based on microservice architecture.



## Design
<!-- The design choices are explained in [design.md](design.md). -->

See the [slides](https://docs.google.com/presentation/d/1nz9gV4Jh8lg4LGen0oAvFr0xh2qfgl8t/) for the presentation.



## Environment Variables
- `ORDER_SERVICE_URL`, `STOCK_SERVICE_URL`, `PAYMENT_SERVICE_URL`: `http://host:port/`  
  The leading protocol and ending slash are necessary.

- `REDIS_ADDRS`: `host:port` or `host:port,host:port,...,host:port`  
  Comma separated list of strings to sharded redis. DO NOT add any space or ending comma.

- `REDIS_PASSWORD`: `pwd`  
  For simplicity, the passwords are required to be the same.

- `REDIS_DB`: `0`  
  For simplicity, the dbs are required to be the same.

- `MACHINE_ID`: in the format of `1/3`  
  Index of machine (i.e. k8s deployment of the same image) slash number of machines. Index starts from 1. Note that the default routing is a simple identical mapping from machine id to redis shard id, thus requiring the number of machines identical to that of redis databases. If they disagree, you should implement your own hashing algorithm.

- `WDM_DEBUG`: `1` or `0`  
  Enable or disable debug logging. Disable logging to improve performance.

- `GIN_MODE`: `debug` or `release`  
  Default to `debug`. Set to `release` for better performance.



## Deployment
You can do either of the followings to deploy the project. We strongly recommend [deploying the project at scale](#deploy-at-scale) to get better performance.


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
Follow this [quickstart](https://cloud.google.com/kubernetes-engine/docs/deploy-app-cluster) up to the section "Create a GKE cluster". Then in the shell execute
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



## Deploy at Scale
As mentioned in the [Design](#design) section, this project can be easily scaled.


### Tested Configuration on GKE
1. Create a **private** (autopilot) cluster, otherwise you may run out of external IPs.

2. Create a cloud NAT gateway to allow external connectivity. In the shell run  
   ```bash
   region=<cluster region>
   gcloud compute addresses create ${region}-nat-ip --region ${region}
   gcloud compute routers create rtr-${region} --network default --region ${region}
   gcloud compute routers nats create nat-gw-${region} --router rtr-${region} --router-region ${region} --region ${region} --nat-external-ip-pool=${region}-nat-ip --nat-all-subnet-ip-ranges
   gcloud compute firewall-rules create all-pods-and-master-ipv4-cidrs --network default --allow all --direction INGRESS --source-ranges 10.0.0.0/8,172.16.2.0/28
   ```

3. Run `echo $(dig +short myip.opendns.com @resolver1.opendns.com)` to reveal the IP of the current shell session. Add it to `Control plane authorized networks` which lies in the details panel of your cluster. Note that you have to do this everytime you start a new shell.

4. Deploy. Run  
   ```bash
   prefix="https://raw.githubusercontent.com/wdm23-5/wdm-project-go/main/"
   curl ${prefix}gke-scale/deploy.sh -o deploy.sh
   source deploy.sh
   ```  
   You may have to wait several minutes after each `helm install` / `kubectl apply` for the services to become stable. Therefore we recommend executing the commands in `deploy.sh` one-by-one by hand.

5. Check the external IP of the ingress. You can access the REST APIs from there.  


As we are poor students, we cannot scale the project too much. Currently, we only have two services and two databases for order / stock / payment microservice respectively. Also, they are limited to using only 1 vCPU each. Such configuration already uses up the default quota.


### I am rich and I want more!
The project can, of course, be scaled to more services / databases / vCPUs, but remember to check your budget and quota first. We expect even better performance if deploying 5+ services and databases per microservice with each utilizing 2+ vCPUs, so that they can run in true parallel. This can be done by increasing the number of deployments in the script. For example,  
```bash
service=<name of service>  # order / stock / payment
for i in {1..5}
do
    echo "--- ${i} ---"
    helm install redis-${service}-${i} bitnami/redis -f gke-scale/helm/redis-helm-values.yaml
    THIS_ID=${i} envsubst < gke-scale/kube/${service}-app.yaml.tpl | kubectl apply -f -
done
```
Do not forget to modify the [env vars](#environment-variables), namely `REDIS_ADDRS` and `MACHINE_ID`, in the corresponding `.yaml.tpl` file before applying. You may also want to increase the amount of CPU (and perhaps memory) resources to be allocated.


### About scaling up on local minikube
Due to time and labour limitation, we have not been able to test whether the `deploy.sh` also works for minikube or not. We are optimistic about the portability of the script itself, but we are not sure if a local machine can provide sufficient hardware resources for so many services.



## Test
Use docker to deploy the project locally and run
```bash
cd test
pip install -r requirements.txt
python test_microservices.py
```
to do the correctness test.

To perform consistency test and stress test, see [wdm-project-benchmark](https://github.com/wdm23-5/wdm-project-benchmark). To get the best performance, consider [deploying the project at scale](#deploy-at-scale).

~~We have a working example following the above [configuration](#tested-configuration-on-gke) available at [34.28.175.41](http://34.28.175.41/) for testing.~~
We ran out of google credits :(



## Contributors
- [Junbo Xiong](https://github.com/C6H5-NO2)
- [Yankun Wang](https://github.com/dear-dd)
- [Yizhen Zhang](https://github.com/zyz5yolo)
