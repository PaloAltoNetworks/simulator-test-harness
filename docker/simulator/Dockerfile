FROM python:3.6-alpine

RUN apk add \
  jq \
  curl \
  bash

RUN apk upgrade

# The default path that will be used for the simulator test.
ENV SIMULATOR_PATH /root/simulator

ENV BIN_PATH /usr/local/bin

ENV APOCTL_DOWNLOAD_URL https://download.aporeto.com/prismacloud/app0/apoctl/linux/apoctl

ENV KUBECTL_VERSION v1.23.2
ENV KUBECTL_DOWNLOAD_URL https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl

ENV GCLOUD_SDK_VERSION 380.0.0

ENV PATH /opt/google-cloud-sdk/bin:$PATH

ENV HELM_VERSION v3.8.1
ENV HELM_DOWNLOAD_URL https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz
ENV HELM_TMP_PATH /opt/helm
ENV HELM_PATH /usr/local/bin/helm

# Install apoctl.
RUN curl -fsSL ${APOCTL_DOWNLOAD_URL} -o ${BIN_PATH}/apoctl \
  && chmod +x ${BIN_PATH}/apoctl

## Install kubectl.
RUN curl -fsSL ${KUBECTL_DOWNLOAD_URL} -o ${BIN_PATH}/kubectl \
  && chmod +x ${BIN_PATH}/kubectl

# Install aws-iam-authenticator.
RUN curl -fsSL -o aws-iam-authenticator https://amazon-eks.s3.us-west-2.amazonaws.com/1.15.10/2020-02-22/bin/linux/amd64/aws-iam-authenticator \
  && chmod +x ./aws-iam-authenticator \
  && mv aws-iam-authenticator /usr/local/bin

# Install helm.
RUN mkdir -p ${HELM_TMP_PATH} \
  && curl -fsSL ${HELM_DOWNLOAD_URL} -o ${HELM_TMP_PATH}/helm.tar.gz \
  && tar -xzf ${HELM_TMP_PATH}/helm.tar.gz -C ${HELM_TMP_PATH} \
  && mv ${HELM_TMP_PATH}/linux-amd64/helm ${HELM_PATH} \
  && chmod +x ${HELM_PATH} \
  && rm -rf ${HELM_TMP_PATH}

RUN curl -fsSL -o /opt/google-cloud-sdk.tar.gz \
  https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
  tar -xzf /opt/google-cloud-sdk.tar.gz -C /opt && \
  rm -f /opt/*.tar.gz

# Copy the binaries into the image.
COPY utils /usr/local/bin

# Copy the charts into the image.
COPY charts ${SIMULATOR_PATH}

# Set the current workdir to the test path.
WORKDIR ${SIMULATOR_PATH}

ENTRYPOINT ["/bin/bash"]
