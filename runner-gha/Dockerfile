# SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26 AS builder
WORKDIR /app

COPY core/go.mod core/go.sum ./core/
COPY runner-gha/go.mod runner-gha/go.sum ./runner-gha/
COPY runner-gitea/go.mod runner-gitea/go.sum ./runner-gitea/
COPY go.work go.work.sum ./
RUN go work sync

COPY . .
ARG GOEXPERIMENT=jsonv2
ARG CGO_ENABLED=0
RUN go build -o ./bin/gha-runner -trimpath -ldflags "-s -w" ./runner-gha

FROM cgr.dev/chainguard/static
COPY --from=builder /app/bin/gha-runner /usr/bin/gha-runner
ENTRYPOINT ["/usr/bin/gha-runner"]
