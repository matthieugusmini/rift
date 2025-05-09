FROM gcr.io/distroless/static
COPY rift /usr/local/bin/rift
ENTRYPOINT [ "/usr/local/bin/rift" ]
