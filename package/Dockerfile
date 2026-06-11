#
# build the target application
#
FROM public.ecr.aws/docker/library/golang:1.26.4-alpine3.24 AS builder

# update the packages
RUN apk update && apk upgrade && apk add --no-cache make

WORKDIR /build
COPY service-common ./service-common
COPY dataverse-connector ./dataverse-connector
RUN cd dataverse-connector; make linux
COPY dspace-connector ./dspace-connector
RUN cd dspace-connector; make linux

#
# build the target container
#
FROM public.ecr.aws/docker/library/alpine:3.24

# update the packages
RUN apk update && apk upgrade && apk add bash tzdata ca-certificates && rm -rf /var/cache/apk/*

# Create the run user and group
RUN addgroup webservice && adduser webservice -G webservice -D

# set the timezone appropriatly
ENV TZ=UTC
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Specify home 
ENV APP_HOME=/aptrust-submit-connectors
WORKDIR ${APP_HOME}

# Create necessary directories
RUN mkdir -p ${APP_HOME} ${APP_HOME}/bin
RUN chown -R webservice ${APP_HOME} && chgrp -R webservice ${APP_HOME}

# Move in necessary assets
COPY package/data/container_bash_profile /home/webservice/.profile
COPY --from=builder /build/dataverse-connector/bin/dataverse-connector.linux ${APP_HOME}/bin/dataverse-connector
COPY --from=builder /build/dspace-connector/bin/dspace-connector.linux ${APP_HOME}/bin/dspace-connector

# Ensure permissions are correct
RUN chown webservice:webservice /home/webservice/.profile ${APP_HOME}/bin/*-connector && chmod 755 /home/webservice/.profile ${APP_HOME}/bin/*-connector

# Add the build tag
ARG BUILD_TAG
RUN test -n "$BUILD_TAG" && touch ${APP_HOME}/buildtag.build-$BUILD_TAG || touch ${APP_HOME}/buildtag.build-0

# Specify the user
USER webservice

#
# end of file
#
