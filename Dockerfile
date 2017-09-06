FROM alpine

RUN adduser -h /home/faasnomad -D faasnomad faasnomad
#RUN usermod -aG docker cnitch

COPY ./faas-nomad /home/faasnomad/
RUN chmod +x /home/faasnomad/faas-nomad

USER faasnomad

ENTRYPOINT ["/home/faasnomad/faas-nomad"]
