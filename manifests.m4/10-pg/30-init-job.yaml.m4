apiVersion: batch/v1
kind: Job
metadata:
  name: init-image-db
spec:
  template:
    spec:
      containers:
      - name: init-image-db
        image: DOCKER_PREFIX/k8s-sl3-pginit:latest
        imagePullPolicy: Always
        env:
        - name: PGUSER
          valueFrom:
            secretKeyRef:
              name: pg-credentials
              key: POSTGRES_USER
        - name: PGPASSWORD
          valueFrom:
            secretKeyRef:
              name: pg-credentials
              key: POSTGRES_PASSWORD
        - name: PGHOST
          value: "pg"
        - name: PGPORT
          value: "5432"
        - name: DB_NAME
          value: "postgres"
        - name: IMAGE_TABLE
          value: "image"        
        - name: IMAGE_NAME_COLUMN
          value: "name"
        - name: IMAGE_URL_COLUMN
          value: "url"
      restartPolicy: Never
