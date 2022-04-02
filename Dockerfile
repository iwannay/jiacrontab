FROM yarnpkg/dev as frontend-env

WORKDIR /jiacrontab
RUN apt-get install git
RUN git clone https://github.com/jiacrontab/jiacrontab-frontend.git
WORKDIR /jiacrontab/jiacrontab-frontend
RUN yarn && yarn build

FROM golang AS jiacrontab-build
WORKDIR /jiacrontab
COPY . .
COPY --from=frontend-env /jiacrontab/jiacrontab-frontend/build /jiacrontab/frontend-build
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct
RUN GO111MODULE=on go install github.com/go-bindata/go-bindata/v3/go-bindata@latest
RUN make build assets=frontend-build

FROM debian AS jiarontab-run
COPY --from=jiacrontab-build /jiacrontab/build /jiacrontab/build
WORKDIR /jiacrontab/bin
VOLUME ["/jiacrontab/bin/data"]
EXPOSE 20001 20000 20003
RUN mv /jiacrontab/build/jiacrontab/jiacrontabd/* . && mv /jiacrontab/build/jiacrontab/jiacrontab_admin/* .
ENTRYPOINT []

