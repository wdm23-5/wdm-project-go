git clone https://github.com/wdm23-5/wdm-project-go.git
cd wdm-project-go

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx -f gke-scale/helm/nginx-helm-values.yaml

helm install redis-order-1 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml
helm install redis-order-2 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml

helm install redis-stock-1 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml
helm install redis-stock-2 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml

helm install redis-payment-1 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml
helm install redis-payment-2 bitnami/redis -f gke-scale/helm/redis-helm-values.yaml

# remember to change the env var if you change the name/number of services (see readme)

THIS_ID=1 envsubst < gke-scale/kube/order-app.yaml.tpl | kubectl apply -f -
THIS_ID=2 envsubst < gke-scale/kube/order-app.yaml.tpl | kubectl apply -f -

THIS_ID=1 envsubst < gke-scale/kube/stock-app.yaml.tpl | kubectl apply -f -
THIS_ID=2 envsubst < gke-scale/kube/stock-app.yaml.tpl | kubectl apply -f -

THIS_ID=1 envsubst < gke-scale/kube/payment-app.yaml.tpl | kubectl apply -f -
THIS_ID=2 envsubst < gke-scale/kube/payment-app.yaml.tpl | kubectl apply -f -

kubectl delete -A ValidatingWebhookConfiguration nginx-ingress-ingress-nginx-admission
kubectl apply -f gke-scale/kube/ingress-service.yaml
