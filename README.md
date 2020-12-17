# Kubernetes skill level 3 solution

## Overview

This repo contains my solution for the Kubernetes skill level 3 exercise for the a8s onboarding.
The task description can be found [here](https://docs.google.com/document/d/1rbtkR2CWnbkSSE5aE_jUUd780y9WC0yxgZ-Lv0ek6Yg/edit#heading=h.1cklqi3vagle).

The solution consists of:

1. A webservice to which images to be blurred (and saved on S3) can be POSTed at the `/picture` path.
2. A worker that downloads images from S3, blurs them and updates them on S3.
3. A RabbitMQ broker that the webservice and the worker use to exchange the references to the images to blur.
4. A PostgreSQL database where metadata about images (name and URL) is saved.

For the webservice (1) and the worker (2), the following are present:

- Go source code.
- Dockerfiles.
- Kubernetes YAML manifests.

For the RabbitMQ and PostgreSQL there's no source code, only Kubernetes YAML manifests are present as they already have official Docker images.

For the PostgreSQL database, a bash script to initialize it (create the table where images metadata will be stored) is provided as well.

**The code quality is pretty poor.
I would have liked to do a second iteration to make it better, but for time reasons I'd rather not do that since this is only a training.** Section [TODOs](#TODOs) also includes some things that would improve code quality.
Also, at least for the worker that blurs images it would have been best to use python over Go as the script with the blurring algorithm is provided in python.

## Limitations

There are also some limitations in functionality:

- This has been tested only with images with `.jpg` extension, please use only that. The main reason
is that the [python script](scripts/imageblur/yolo_opencv.py) that blurs images has only that format hardcoded (although it would not be too difficult to change it to make it more flexible).
- The webservice understands only multi-part/form-data requests (see [here](https://stackoverflow.com/questions/16958448/what-is-http-multipart-request) and [here](https://www.w3.org/TR/html401/interact/forms.html#h-17.13.4.2) to learn about those) where the picture part has name "image", so only stick to those requests. An example of one such request is in section [Test](#Test).
- Neither RabbitMQ nor PostgreSQL are HA. An **optional** part of the exercise was to make RabbitMQ HA. I'd rather not do that for time reasons.

## A tour of the folders and files

### Source code folders

- [cmd/](cmd/): Go main packages for the webservice and the worker.
- [pkg/](pkg/): Root of all Go source code with the exception of the main packages.
- [pkg/webservice/](pkg/webservice/): Code for the webservice.
- [pkg/imageblur/](pkg/imageblur/): Code for the worker that blurs images.
- [pkg/helper/](pkg/helper/): Root of helper code to interact with AWS S3, RabbitMQ, and PostgreSQL.
- [pkg/helper/s3/](pkg/helper/s3/): Code to interact with AWS S3.
- [pkg/helper/rabbitmq/](pkg/helper/rabbitmq/): Code to interact with RabbitMQ.
- [pkg/helper/pg/](pkg/helper/pg/): Code to interact with PostgreSQL.

### Helper scripts

- [scripts/imageblur](scripts/imageblur): `python` `OpenCV` script and ancillary files that perform the filtering that blurs images (this is called from the Go code of the worker that blurs images).
- [scripts/pg](scripts/pg): `bash` script that uses `psql` to create the table where images metadata is stored inside the PostgreSQL DB; it is meant to run inside a Kuberentes Job.

### Dockerfiles

- [images/](images/): Root dir where all Dockerfiles are.
- [images/webservice/](images/webservice/): Contains the Dockerfile for the webservice
- [images/imageblur/](images/imageblur/): Contains the Dockerfile for the imageblurring worker
- [images/pginit/](images/pginit/): Contains the Dockerfile for a script that initializes PostgreSQL by creating the table where the images metadata is stored.

### YAML manifests

- [manifests/10-rabbitmq](manifests/10-rabbitmq): All manifests for RabbitMQ.
- [manifests/10-pg](manifests/10-pg): All manifests for PostgreSQL.
- [manifests/20-webservice](manifests/20-webservice): All manifests for the webservice.
- [manifests/20-imageblur](manifests/20-imageblur): All manifests for the worker that blurs images.
- [manifests.m4/](manifests.m4/): [m4](https://en.wikipedia.org/wiki/M4_(computer_language)) templates for the manifests that use Docker images that are owned by this project.

More precisely, the manifests in [manifests/](manifests/) only reference Docker images on my Dockerhub. You on the other hand, might want to build and publish Docker images for the components in this repo to your Dockerhub and then deploy everything on Kubernetes using those images, and you might accomplish this by using the templates in [manifests.m4](manifests.m4/) (indirectly via `make`, more details are in the sections [Build](#Build) and [Deploy](#Deploy)).

Since this project is about blurring images, there are also two, example images under [example-pictures/](example-pictures/) that you can use to test the project.

## Automation 

There's a [Makefile](Makefile) to build and deploy everything.
It defines several variables that you can set to alter the default behavior, but for which defaults are provided (so you don't have to set them):

- `DOCKER_PREFIX`: Dockerhub account to which container images are pushed, defaults to your login name on the machine where you cloned this repo.
- `GOOS`: OS for which the Go source code is compiled, defaults to `linux`.
- `GOARCH`: arch for which the Go source code is compiled, defaults to `amd64`.
- `POSTGRES_USER`: user to configure in the PostgreSQL cluster that the webservice will use, defaults to `user`.
- `POSTGRES_PASSWORD`: password for the `POSTGRES_USER`, defaults to `password`.
- `RABBITMQ_USER`: user to configure in the RabbitMQ broker, that the webservice and the image blurring worker use, defaults to `user`.
- `RABBITMQ_PASSWORD`: password for the `RABBITMQ_USER`, defaults to `password`.
- `RABBITMQ_QUEUE`: name of the queue that the webservice and the image blurring worker will use. Defaults to `images`. The webservice and the image blurring worker will create this queue if it does not exist.

There are also other variables that will be described where more appropriate.

### Build

- build Go source code: `$ make go-build`
- build Docker container images: `$ make go-build docker-build`
- push Docker container images to Dockerhub: `$ make go-build docker-build docker-push`

Notice that if you want to push to your Dockerhub and your username there does not
match your login name on your machine, the syntax changes to:

```
make DOCKER_PREFIX=<your-dockerhub-name> go-build docker-build docker-push
```

### Deploy and undeploy

The targets described in this section assume that you have `kubectl` installed and configured to point at a Kubernetes cluster.

The makefile has the `deploy` and `undeploy` targets, to deploy and undeploy everything on the Kubernetes cluster.

**Because the webservice and the worker upload and download pictures from an `AWS S3 bucket`, to run the `deploy` target you MUST set variables that identify/configure that bucket, otherwise the `deploy` target will fail.** Obviously, this also means that you need to have an S3 bucket on AWS.

The required variables are:

- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `AWS_REGION`
- `S3_BUCKET`

So the syntax is:

```
make AWS_ACCESS_KEY=<key> ... S3_BUCKET=<bucket> deploy
```

`make deploy` will use the manifests in [manifests/](mainfests/) which use the Docker container
images in my Dockerhub. If you have built and pushed the Docker images to your Dockerhub and want to use those instead,
you'll have to update the manifests that reference container images created in this repo to use your Dockerhub account name. To do so, you'll have to run the following command:

```
make DOCKER_PREFIX=<your-dockerhub-name> manifests/10-pg/30-init-job.yaml manifests/20-webservice/20-deployment.yaml manifests/20-imageblur/20-deployment.yaml
```

## Test

First set up port forwarding between your machine and the webservice in the Kubernetes cluster to be able to send requests to the webservice:

```
kubectl port-forward service/webservice 8080:8080
```

Then, send a request:

```
curl -X POST -H "Content-Type: multipart/form-data" -F "image=@example-pictures/bill-gates.jpg" http://localhost:8080/picture
```

If you then go to the AWS S3 bucket you configured when you deployed everything, you should see that
it stores the image [bill-gates.jpg](example-pictures/bill-gates.jpg), where the face in the picture is blurred.
You can use any picture you want, as long as its extension is `.jpg`.

Also, for any picture that you send, you should see an entry with the name of that picture and the URL of that picture in S3 in the table called `IMAGE` inside the PostgreSQL database.

## TODOs (that we are not going to do)

- Make handling of images idempotent (or add clean-up-on-error logic). Currently, insertion on PostgreSQL is not idempotent. So, if for instance, the webservice crashes after inserting metadata for a picture in PostgreSQL but before enqueueing the picture URL in RabbitMQ, following attempts to process the picture will fail.
- Make logging non-embarassing (not enough logs, terrible library used, useless log messages).
- Add/improve HTTP error messages returned to the user.
- Add tests.
- Test more. Only the happy path and a few error paths have been tested, definitely not enough.
- Make webservice and worker code independent from specific metadata DB/object store/message broker. For example, the webservice code should not directly use RabbitMQ. It should only see a generic interface to a message broker, and the actual message broker implementation should be pluggable (something akin to Dependency Injection).
- Harmonize code in [pkg/s3/](pkg/s3/). That's code that should be symmetric by nature. For example, there's a function to upload images and one to donwload images. Such functions should have symmetric signatures as this improves code quality, instead they have completely different signatures. For instance, one gets passed some state as an argument, while the other internally creates the same state.
- [pkg/s3/](pkg/s3/): switch to reusing the same AWS S3 session when the worker that blurs images processes an image. That worker first downloads the image from S3, blurs it, and re-uploads it on S3. Currently two different sessions are created, one to download and one to upload. That's stupid, there should be a unique one.
- Make fields configurable: a LOT of fields are hardcoded and should be configurable instead (e.g. RabbitMQ server DNS name)
- Currently only ENV vars are supported for configuration, add support for config files as well.

There are many more than I thought about at some point but that I cannot remember right now. Most of them are described as comments near the relevant code.
