FROM alpine:latest

RUN adduser -h /home/faasnomad -D faasnomad faasnomad

COPY faas-nomad /home/faasnomad/
RUN chmod +x /home/faasnomad/faas-nomad

USER faasnomad

ENTRYPOINT ["/home/faasnomad/faas-nomad"]
