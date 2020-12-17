apiVersion: apps/v1
kind: Deployment
metadata:
  name: imageblur
  labels:
    role: imageblur
    exercise: k8s-skill-lvl-3
spec:
  replicas: 2
  selector:
    matchLabels:
      role: imageblur
      exercise: k8s-skill-lvl-3
  template:
    metadata:
      labels:
        role: imageblur
        exercise: k8s-skill-lvl-3
    spec:
      containers:
      - name: imageblur
        image: DOCKER_PREFIX/k8s-sl3-imageblur:latest
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
      