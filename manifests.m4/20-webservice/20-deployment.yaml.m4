apiVersion: apps/v1
kind: Deployment
metadata:
  name: webservice
  labels:
    role: webservice
    exercise: k8s-skill-lvl-3
spec:
  replicas: 2
  selector:
    matchLabels:
      role: webservice
      exercise: k8s-skill-lvl-3
  template:
    metadata:
      labels:
        role: webservice
        exercise: k8s-skill-lvl-3
    spec:
      containers:
      - name: webservice
        image: DOCKER_PREFIX/k8s-sl3-webservice:latest
        imagePullPolicy: Always
        env:
        - name: RABBITMQ_USER 
          valueFrom:
            secretKeyRef:
              name: rabbitmq-credentials
              key: RABBITMQ_USER
        - name: RABBITMQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rabbitmq-credentials
              key: RABBITMQ_PASSWORD
        - name: RABBITMQ_QUEUE
          valueFrom:
            secretKeyRef:
              name: rabbitmq-credentials
              key: RABBITMQ_QUEUE
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: pg-credentials
              key: POSTGRES_USER
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: pg-credentials
              key: POSTGRES_PASSWORD
        - name: AWS_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: AWS_ACCESS_KEY
        - name: AWS_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: AWS_SECRET_KEY
        - name: AWS_REGION
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: AWS_REGION
        - name: S3_BUCKET
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: S3_BUCKET
      