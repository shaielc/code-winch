FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /out/winch-sandbox ./cmd/winch-sandbox \
 && CGO_ENABLED=0 go build -o /out/winch ./cmd/winch \
 && CGO_ENABLED=0 go build -o /out/fake-harness ./cmd/fake-harness

FROM alpine:3.22
RUN addgroup -S winch && adduser -S -G winch -u 10001 winch
COPY --from=go-build /out/ /usr/local/bin/
COPY --from=web /src/web/dist/ /opt/winch/web/
USER winch
EXPOSE 8080
ENTRYPOINT ["winch-sandbox"]
