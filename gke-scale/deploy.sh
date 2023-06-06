helm install nginx-ingress ingress-nginx/ingress-nginx

helm install redis-order-1 bitnami/redis -f ${prefix}gke-scale/helm/redis-helm-values.yaml
helm install redis-order-2 bitnami/redis -f ${prefix}gke-scale/helm/redis-helm-values.yaml
helm install redis-order-3 bitnami/redis -f ${prefix}gke-scale/helm/redis-helm-values.yaml

helm install redis-stock-1 bitnami/redis -f ${prefix}gke-scale/helm/redis-helm-values.yaml

helm install redis-payment-1 bitnami/redis -f ${prefix}gke-scale/helm/redis-helm-values.yaml

curl -s ${prefix}gke-scale/kube/order-app.yaml.tpl | THIS_ID=1 envsubst | kubectl apply -f -
curl -s ${prefix}gke-scale/kube/order-app.yaml.tpl | THIS_ID=2 envsubst | kubectl apply -f -
curl -s ${prefix}gke-scale/kube/order-app.yaml.tpl | THIS_ID=3 envsubst | kubectl apply -f -

curl -s ${prefix}gke-scale/kube/stock-app.yaml.tpl | THIS_ID=1 envsubst | kubectl apply -f -

curl -s ${prefix}gke-scale/kube/payment-app.yaml.tpl | THIS_ID=1 envsubst | kubectl apply -f -

kubectl apply -f ${prefix}gke-scale/kube/ingress-service.yaml
