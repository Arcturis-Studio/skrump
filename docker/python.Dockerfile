# Credit: https://github.com/alexdmoss/distroless-python
## -------------- Layer to give access to newer python + its dependencies ------------- ##
### Base image for Python 3.12 with slim bookworm optimization ###
FROM python:3.12-slim-bookworm AS python-base

## ------------ Setup standard non-root user for use downstream ------------##
### Create a standard non-root user for running applications ###
ARG NONROOT_USER="nonroot"
ARG NONROOT_GROUP="nonroot"

RUN groupadd ${NONROOT_GROUP} \
    && useradd -m ${NONROOT_USER} -g ${NONROOT_GROUP}

USER ${NONROOT_USER}

ENV PATH="/home/${NONROOT_USER}/.local/bin:${PATH}"

## ------------ Setup user environment with good python practices ------------##
### Set up a Python development environment with best practices ###
USER ${NONROOT_USER}
WORKDIR /home/${NONROOT_USER}

# Set local, stop '.pyc' generation, and enable tracebacks on segfaults
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONFAULTHANDLER 1

## ------------- pipenv/poetry for use elsewhere as builder image ------------ ##
### Install pipenv, poetry, and virtualenv for package management ###
RUN pip install --upgrade pip && \
    pip install --no-warn-script-location virtualenv poetry pipenv

## ------------------------------- Distroless base image ------------------------------##
### Use the distroless C or cc:debug image as a base for building ###
FROM gcr.io/distroless/cc-debian12 AS runner-image

## ------------------------- Copy python itself from builder --------------------------##
### Copy Python and its dependencies to this layer, reducing image size ###
COPY --from=python-base /usr/local/lib/ /usr/local/lib/
COPY --from=python-base /usr/local/bin/python /usr/local/bin/
COPY --from=python-base /etc/ld.so.cache /etc/

## -------------------------- Add common compiled libraries ---------------------------##
### Copy necessary system libraries to this layer, avoiding import errors ###
ARG CHIPSET_ARCH

# Required by lots of packages - e.g. six, numpy, wsgi
COPY --from=python-base /lib/${CHIPSET_ARCH}-linux-gnu/libz.so.1 /lib/${CHIPSET_ARCH}-linux-gnu/libselinux.so.1 /lib/${CHIPSET_ARCH}-linux-gnu/
# Required by google-cloud/grpcio
COPY --from=python-base /usr/lib/${CHIPSET_ARCH}-linux-gnu/libffi* /usr/lib/${CHIPSET_ARCH}-linux-gnu/
COPY --from=python-base /lib/${CHIPSET_ARCH}-linux-gnu/libexpat* /lib/${CHIPSET_ARCH}-linux-gnu/
# Required for job dependencies
COPY --from=python-base /usr/local/bin/pip /usr/local/bin/

## --------------------------- Add shell ---------------------------- ##
### Copy common system binaries to this layer ###
COPY --from=python-base /bin/echo /bin/ln /bin/rm /bin/sh /bin/

## -------------------------------- Non-root user setup -------------------------------##
### Set up the non-root user for running applications ###
ARG NONROOT_USER="nonroot"
ARG NONROOT_GROUP="nonroot"

# Quick validation that python still works whilst we have a shell
# pipenv links python to python3 in venv
RUN echo "${NONROOT_USER}:x:1000:${NONROOT_GROUP}" >> /etc/group \
    && echo "${NONROOT_USER}:x:1001:" >> /etc/group \
    && echo "${NONROOT_USER}:x:1000:1001::/home/${NONROOT_USER}:" >> /etc/passwd \
    && python --version \
    && ln -s /usr/local/bin/python /usr/local/bin/python3 \
    && rm /bin/echo /bin/ln /bin/rm

## --------------------------- Standardise execution env -----------------------------##
### Set the default environment for running applications ###
# Default to running as non-root, uid=1000
USER ${NONROOT_USER}

# Standardise on locale, don't generate .pyc, enable tracebacks on seg faults
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONFAULTHANDLER 1

# ENTRYPOINT ["/usr/local/bin/python"]
ENTRYPOINT ["/bin/sh"]
