# SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24 AS builder
ENV CGO_ENABLED=0
COPY core/go.mod core/go.sum ./core/
COPY runner-gha/go.mod runner-gha/go.sum ./runner-gha/
COPY runner-gitea/go.mod runner-gitea/go.sum ./runner-gitea/
COPY go.work go.work.sum ./
RUN go work sync
COPY . .
RUN go build -o /usr/bin/gha-runner -trimpath -ldflags "-s -w" ./runner-gha

FROM cgr.dev/chainguard/static
COPY --from=builder /usr/bin/gha-runner /usr/bin/gha-runner
ENTRYPOINT ["/usr/bin/gha-runner"]
