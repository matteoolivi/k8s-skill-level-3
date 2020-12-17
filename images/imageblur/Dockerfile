FROM ubuntu:18.04

ENV TZ=Europe/Berlin
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update && apt-get install -y \
    python3 \
    python3-numpy \
    python3-setuptools \
    python3-pip \
    libopencv-dev \
    opencv-data

COPY scripts/imageblur/* /scripts/imageblur/

RUN pip3 install --upgrade pip
RUN pip3 install opencv-python

COPY bin/imageblur /imageblur
CMD [ "/imageblur" ]
