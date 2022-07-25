FROM alpine

RUN mkdir /app
ADD ./mysql-proxy /app/
ADD ./conf.toml  /app/
WORKDIR /app

CMD ["./mysql-proxy"]