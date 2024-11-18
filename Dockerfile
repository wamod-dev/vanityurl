FROM scratch

ENV VANITYURL_CONFIG=/etc/vanityurl/config.yml

COPY vanityurl /bin/vanityurl

ENTRYPOINT ["/bin/vanityurl"]
