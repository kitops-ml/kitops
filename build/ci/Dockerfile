FROM bats/bats

WORKDIR /app

# copy script under test and tests
COPY ./dockerfiles/KServe/entrypoint.sh ./entrypoint.sh
COPY tests ./tests

# ensure entrypoint is executable
RUN chmod +x entrypoint.sh

ENTRYPOINT ["bats", "tests"]