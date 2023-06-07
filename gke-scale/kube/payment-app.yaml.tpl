apiVersion: apps/v1
kind: Deployment
metadata:
  name: payment-deployment-${THIS_ID}
spec:
  replicas: 1
  selector:
    matchLabels:
      component: payment-${THIS_ID}
  template:
    metadata:
      labels:
        app: payment-app
        component: payment-${THIS_ID}
    spec:
      containers:
        - name: payment-${THIS_ID}
          image: ghcr.io/wdm23-5/wdm-project-go/payment:latest
          resources:
            limits:
              memory: "2Gi"
              cpu: "1"
            requests:
              memory: "1Gi"
              cpu: "1"
          command: ["./payment-gin"]
          args: [""]
          ports:
            - containerPort: 5000
          env:
            - name: REDIS_ADDRS
              value: "redis-payment-1-master:6379,redis-payment-2-master:6379"
            - name: REDIS_PASSWORD
              value: "redis"
            - name: REDIS_DB
              value: "0"
            - name: ORDER_SERVICE_URL
              value: "http://nginx-ingress-ingress-nginx-controller.default.svc.cluster.local:80/orders/"
            - name: MACHINE_ID
              value: "${THIS_ID}/2"
            - name: WDM_DEBUG
              value: "0"
            - name: GIN_MODE
              value: "release"
