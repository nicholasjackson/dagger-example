FROM alpine:latest

RUN mkdir /app
RUN ls -las
COPY ./build/dagger-example /app/dagger-example
RUN chmod +x /app/dagger-example

ENTRYPOINT ["/app/dagger-example"]
