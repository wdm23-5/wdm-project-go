apiVersion: v1
kind: Service
metadata:
  name: stock-service-${THIS_ID}
spec:
  type: ClusterIP
  selector:
    component: stock-${THIS_ID}
  ports:
    - port: 5000
      name: http
      targetPort: 5000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stock-deployment-${THIS_ID}
spec:
  replicas: 1
  selector:
    matchLabels:
      component: stock-${THIS_ID}
  template:
    metadata:
      labels:
        component: stock-${THIS_ID}
    spec:
      containers:
        - name: stock-${THIS_ID}
          image: ghcr.io/wdm23-5/wdm-project-go/stock:latest
          resources:
            limits:
              memory: "1Gi"
              cpu: "1"
            requests:
              memory: "1Gi"
              cpu: "1"
          command: ["./stock-gin"]
          args: [""]
          ports:
            - containerPort: 5000
          env:
            - name: REDIS_ADDRS
              value: "redis-stock-1-master:6379"
            - name: REDIS_PASSWORD
              value: "redis"
            - name: REDIS_DB
              value: "0"
            - name: MACHINE_ID
              value: "${THIS_ID}/1"
            - name: WDM_DEBUG
              value: "0"
            - name: GIN_MODE
              value: "release"
