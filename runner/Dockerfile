ARG RUNNER_IMAGE=ghcr.io/actions/actions-runner:latest
ARG NODE_IMAGE=node:22-bookworm-slim

FROM ${NODE_IMAGE} AS node-runtime
FROM ${RUNNER_IMAGE}

USER root

ARG CODEX_VERSION=latest

COPY --from=node-runtime /usr/local/ /usr/local/

RUN apt-get update \
    && apt-get install --yes --no-install-recommends jq python3 \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /home/runner/.codex /home/runner/_work /var/lib/code-winch \
    && chown -R runner:runner /home/runner/.codex /home/runner/_work /var/lib/code-winch
RUN npm install --global "@openai/codex@${CODEX_VERSION}"

COPY --chown=runner:runner runner/entrypoint.sh /usr/local/bin/code-winch-runner

USER runner
WORKDIR /home/runner

ENTRYPOINT ["/usr/local/bin/code-winch-runner"]
