apiVersion: v1
kind: Service
metadata:
  name: order-service-${THIS_ID}
spec:
  type: ClusterIP
  selector:
    component: order-${THIS_ID}
  ports:
    - port: 5000
      name: http
      targetPort: 5000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-deployment-${THIS_ID}
spec:
  replicas: 1
  selector:
    matchLabels:
      component: order-${THIS_ID}
  template:
    metadata:
      labels:
        component: order-${THIS_ID}
    spec:
      containers:
        - name: order-${THIS_ID}
          image: ghcr.io/wdm23-5/wdm-project-go/order:latest
          resources:
            limits:
              memory: "2Gi"
              cpu: "2"
            requests:
              memory: "2Gi"
              cpu: "2"
          command: ["./order-gin"]
          args: [""]
          ports:
            - containerPort: 5000
          env:
            - name: REDIS_ADDRS
              value: "redis-order-1-master:6379,redis-order-2-master:6379,redis-order-3-master:6379"
            - name: REDIS_PASSWORD
              value: "redis"
            - name: REDIS_DB
              value: "0"
            - name: PAYMENT_SERVICE_URL
              value: "http://nginx-ingress-ingress-nginx-controller.default.svc.cluster.local:80/payment/"
            - name: STOCK_SERVICE_URL
              value: "http://nginx-ingress-ingress-nginx-controller.default.svc.cluster.local:80/stock/"
            - name: MACHINE_ID
              value: "${THIS_ID}/3"
            - name: WDM_DEBUG
              value: "0"
            - name: GIN_MODE
              value: "release"
