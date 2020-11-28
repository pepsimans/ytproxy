FROM heroku/heroku:20-build as build

COPY . /app
WORKDIR /app

RUN mkdir -p /tmp/buildpack/heroku/go /tmp/build_cache /tmp/env

RUN curl https://codon-buildpacks.s3.amazonaws.com/buildpacks/heroku/go.tgz | tar xz -C /tmp/buildpack/heroku/go && \
    curl -L https://youtube-dl.org/downloads/latest/youtube-dl -o youtube-dl && chmod a+rx youtube-dl

# Execute Buildpack
RUN STACK=heroku-20 /tmp/buildpack/heroku/go/bin/compile /app /tmp/build_cache /tmp/env

# Prepare final image
FROM heroku/heroku:20

COPY --from=build /app/youtube-dl /app/youtube-dl
COPY --from=build /app/bin/proxy /app/proxy
ENV HOME /app

WORKDIR /app
RUN useradd -m heroku
USER heroku
CMD "/app/proxy"