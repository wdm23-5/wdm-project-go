helm install nginx-ingress ingress-nginx/ingress-nginx

helm install redis-order-1 bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml
helm install redis-order-2 bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml
helm install redis-order-3 bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml

helm install redis-stock-1 bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml

helm install redis-payment-1 bitnami/redis -f ${prefix}gke/helm/redis-helm-values.yaml

THIS_ID=1 envsubst < ${prefix}gke/kube/order-app.yaml.tpl | kubectl apply -f -
THIS_ID=2 envsubst < ${prefix}gke/kube/order-app.yaml.tpl | kubectl apply -f -
THIS_ID=3 envsubst < ${prefix}gke/kube/order-app.yaml.tpl | kubectl apply -f -

THIS_ID=1 envsubst < ${prefix}gke/kube/stock-app.yaml.tpl | kubectl apply -f -

THIS_ID=1 envsubst < ${prefix}gke/kube/payment-app.yaml.tpl | kubectl apply -f -

kubectl apply -f ${prefix}gke/kube/ingress-service.yaml
