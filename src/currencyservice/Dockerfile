FROM node:12-alpine as base

FROM base as builder

# Some packages (e.g. @google-cloud/profiler) require additional
# deps for post-install scripts
RUN apk add --update --no-cache \
    python3 \
    make \
    g++

WORKDIR /usr/src/app

COPY package*.json ./

# RUN npm config set registry http://mirrors.cloud.tencent.com/npm/
RUN npm config set registry https://registry.npmjs.org/

RUN npm install --only=production

FROM base

# RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
#     wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#     chmod +x /bin/grpc_health_probe
COPY grpc_health_probe-linux-amd64 /bin/grpc_health_probe

RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0  && \
    chmod +x /bin/grpc_health_probe

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/node_modules ./node_modules

COPY . .

EXPOSE 7000

# ENTRYPOINT [ "node", "server.js" ]
CMD [ "node", "server.js" ]
