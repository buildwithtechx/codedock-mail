FROM node:22-alpine AS admin-builder

WORKDIR /app

COPY package.json package-lock.json ./
COPY apps/admin/package.json ./apps/admin/
RUN npm ci --workspace=@codedock-mail/admin

COPY apps/admin/ ./apps/admin/
COPY tsconfig.base.json ./

RUN npm run build:admin

FROM golang:1.26-alpine AS daemon-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY apps/admin/*.go ./apps/admin/
COPY --from=admin-builder /app/apps/admin/dist ./apps/admin/dist

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /maild ./cmd/maild

FROM alpine:3.21

RUN apk add --no-cache ca-certificates postgresql-client

COPY --from=daemon-builder /maild /maild

EXPOSE 8081

ENTRYPOINT ["/maild"]
