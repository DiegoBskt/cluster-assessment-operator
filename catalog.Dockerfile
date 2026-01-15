# Build the File Based Catalog (FBC) image
# Uses community opm image that supports multi-arch (amd64/arm64)
#
# NOTE: Cache is generated at runtime to avoid QEMU emulation issues
# when building amd64 images on Apple Silicon (arm64) machines.

ARG OCP_VERSION=v4.14
ARG BASE_IMAGE=quay.io/operator-framework/opm:latest

FROM ${BASE_IMAGE}

ARG OCP_VERSION
ARG OPERATOR_NAME=cluster-assessment-operator

# Copy catalog configuration
COPY catalogs/${OCP_VERSION}/${OPERATOR_NAME} /configs/${OPERATOR_NAME}

# Labels
LABEL operators.operatorframework.io.index.configs.v1=/configs
LABEL name="cluster-assessment-operator-catalog"
LABEL vendor="Community"
LABEL version="${OCP_VERSION}"
LABEL summary="FBC catalog for Cluster Assessment Operator"

EXPOSE 50051

ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs"]
