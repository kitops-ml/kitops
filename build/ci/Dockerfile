FROM bats/bats

WORKDIR /app

# copy script under test and tests
COPY ./dockerfiles/KServe/entrypoint.sh ./entrypoint.sh
COPY tests ./tests

# ensure entrypoint is executable and tests dir is writable
RUN chmod +x entrypoint.sh && \
    chown -R nobody:nogroup /app

USER nobody

ENTRYPOINT ["bats", "tests"]
