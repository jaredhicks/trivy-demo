# syntax=docker/dockerfile:1

FROM golang:1.19 as builder
RUN useradd -u 1001 -m iamuser

RUN curl -ksSL https://gitlab.mitre.org/mitre-scripts/mitre-pki/raw/master/os_scripts/install_certs.sh | MODE=debian sh
COPY . /src
WORKDIR /src

RUN go mod download && go mod verify
RUN go build -o /windsvc


FROM scratch
MAINTAINER "TARGETS Team" <jahicks@mitre.org>

COPY --from=builder /windsvc /windsvc
COPY --from=builder /etc/passwd /etc/passwd

EXPOSE 8080
ENTRYPOINT ["/windsvc"]
