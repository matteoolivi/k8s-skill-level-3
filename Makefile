# Default name of the dockerhub account where to publish images.
DOCKER_PREFIX=${LOGNAME}

# Default OS and Arch for which the Go source code in this repo should be built.
GOOS=linux
GOARCH=amd64

# Default credentials for the Postgres DB where image names and URLs are stored.
POSTGRES_USER=user
POSTGRES_PASSWORD=password

# Default credentials for the RabbitMQ broker where image URLs are enqueued.
RABBITMQ_USER=user
RABBITMQ_PASSWORD=password
RABBITMQ_QUEUE=images

docker-push:: docker-push-webservice
docker-push:: docker-push-imageblur
docker-push:: docker-push-pginit

docker-build:: docker-build-webservice
docker-build:: docker-build-imageblur
docker-build:: docker-build-pginit

go-build:: go-build-webservice
go-build:: go-build-imageblur

clean:
	rm -rf bin/

# Targets to compile go code.

go-build-webservice:
	CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=${GOOS} \
	go build -a -o bin/webservice github.com/matteoolivi/img-blurring-exercise/cmd/webservice

go-build-imageblur:
	CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=${GOOS} \
	go build -a -o bin/imageblur github.com/matteoolivi/img-blurring-exercise/cmd/imageblur

# Targets to build docker images.

docker-build-webservice:
	docker build -t ${DOCKER_PREFIX}/k8s-sl3-webservice:latest -f images/webservice/Dockerfile .

docker-build-imageblur: scripts/imageblur/yolov3.weights
	docker build -t ${DOCKER_PREFIX}/k8s-sl3-imageblur:latest -f images/imageblur/Dockerfile .

docker-build-pginit: scripts/pg/init-db.sh
	docker build -t ${DOCKER_PREFIX}/k8s-sl3-pginit:latest -f images/pginit/Dockerfile .

# Targets to publish docker images.

docker-push-webservice:
	docker push ${DOCKER_PREFIX}/k8s-sl3-webservice:latest

docker-push-imageblur:
	docker push ${DOCKER_PREFIX}/k8s-sl3-imageblur:latest

docker-push-pginit:
	docker push ${DOCKER_PREFIX}/k8s-sl3-pginit:latest

.PHONY: manifests/20-webservice/20-deployment.yaml
manifests/10-pg/30-init-job.yaml:
	rm -f manifests/10-pg/30-init-job.yaml && \
	m4 -DDOCKER_PREFIX=${DOCKER_PREFIX} manifests.m4/10-pg/30-init-job.yaml.m4 > manifests/10-pg/30-init-job.yaml

.PHONY: manifests/20-webservice/20-deployment.yaml
manifests/20-webservice/20-deployment.yaml:
	rm -f manifests/20-webservice/20-deployment.yaml && \
	m4 -DDOCKER_PREFIX=${DOCKER_PREFIX} manifests.m4/20-webservice/20-deployment.yaml.m4 > manifests/20-webservice/20-deployment.yaml

.PHONY: manifests/20-imageblur/20-deployment.yaml
manifests/20-imageblur/20-deployment.yaml:
	rm -f manifests/20-imageblur/20-deployment.yaml && \
	m4 -DDOCKER_PREFIX=${DOCKER_PREFIX} manifests.m4/20-imageblur/20-deployment.yaml.m4 > manifests/20-imageblur/20-deployment.yaml

# Targets to ensure that the variables that are needed but for which no default can be provided
# were set by the caller.
# TODO: See if we can use only one `ifndef` by using implicit targets.

aws-credentials-or-die:
ifndef AWS_ACCESS_KEY
	$(error AWS_ACCESS_KEY is undefined)
endif
ifndef AWS_SECRET_KEY
	$(error AWS_SECRET_KEY is undefined)
endif
ifndef AWS_REGION
	$(error AWS_REGION is undefined)
endif
ifndef S3_BUCKET
	$(error S3_BUCKET is undefined)
endif

# Targets to (un)deploy all the components on K8s.
# TODO: These targets should be broken down into smaller ones.

.PHONY: deploy
deploy: aws-credentials-or-die manifests/10-pg/30-init-job.yaml
	kubectl create secret generic rabbitmq-credentials \
		--from-literal=RABBITMQ_USER=${RABBITMQ_USER} \
		--from-literal=RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD} \
		--from-literal=RABBITMQ_QUEUE=${RABBITMQ_QUEUE} && \
	kubectl create secret generic pg-credentials \
		--from-literal=POSTGRES_USER=${POSTGRES_USER} \
		--from-literal=POSTGRES_PASSWORD=${POSTGRES_PASSWORD} && \
	kubectl create secret generic aws-credentials \
		--from-literal=AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
		--from-literal=AWS_SECRET_KEY=${AWS_SECRET_KEY} \
		--from-literal=AWS_REGION=${AWS_REGION} \
		--from-literal=S3_BUCKET=${S3_BUCKET} && \
	kubectl apply -f manifests/10-rabbitmq && \
	kubectl apply -f manifests/10-pg/10-service.yaml && \
	kubectl apply -f manifests/10-pg/20-statefulset.yaml && \
	while kubectl get --no-headers pod pg-0 | grep -vi running > /dev/null; \
		do echo waiting for pod pg-0 to be ready... && sleep 5; done && \
	kubectl apply -f manifests/10-pg/30-init-job.yaml && \
	kubectl apply -f manifests/20-webservice/ && \
	kubectl apply -f manifests/20-imageblur/

.PHONY: undeploy
undeploy: manifests/10-pg/30-init-job.yaml
	kubectl delete --ignore-not-found -f manifests/20-imageblur/ && \
	kubectl delete --ignore-not-found -f manifests/20-webservice/ && \
	kubectl delete --ignore-not-found -f manifests/10-rabbitmq/ && \
	kubectl delete --ignore-not-found -f manifests/10-pg/ && \
	kubectl delete --ignore-not-found pvc -l role=pictures-queue,exercise=k8s-skill-lvl-3  && \
	kubectl delete --ignore-not-found pvc -l role=pictures-metadata-db,exercise=k8s-skill-lvl-3  && \
	kubectl delete --ignore-not-found secret aws-credentials
	kubectl delete --ignore-not-found secret rabbitmq-credentials
	kubectl delete --ignore-not-found secret pg-credentials

# Target to get weights for the image blurring algorithm.

scripts/imageblur/yolov3.weights:
	wget https://pjreddie.com/media/files/yolov3.weights -O scripts/imageblur/yolov3.weights
