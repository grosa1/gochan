FROM golang:1.16-alpine3.15

RUN apk --no-cache add \
						postgresql-client	\
						mariadb-client		\
						nginx 				\
						ffmpeg				\
						python3				\
						git					\
						gcc					\
						musl-dev			\
						openssl				\
						bash

RUN mkdir -p /root/bin \
	&& ln -s /usr/lib/go-1.16/bin/* /root/bin/ \
	&& export PATH=$PATH:/root/bin \
	&& echo "export PATH=$PATH:/root/bin" >> /root/.bashrc \
	&& rm -f /etc/nginx/sites-enabled/* /etc/nginx/sites-available/* \
	&& mkdir -p /var/lib/nginx \
	&& mkdir -p /var/lib/nginx/tmp \
	&& mkdir -p /run/nginx/

WORKDIR /opt/gochan

# Get dependencies
COPY build.py .
RUN ./build.py dependencies

RUN rm -f /etc/nginx/http.d/default.conf
COPY sample-configs/gochan-fastcgi.nginx /etc/nginx/http.d/gochan.conf
COPY sample-configs/gochan.example.json /etc/gochan/gochan.json

# Get all
COPY . .

EXPOSE 9000

# The openssl command will generate self-signed certificate since some browsers like
# Firefox and Chrome automatically do HTTPS requests. this will likely show a warning in
# the browser, which you can ignore
CMD ls -la /opt/gochan && ls -la && ls -la .. && sed -i /etc/gochan/gochan.json \
	-e 's/"Port": 8080/"Port": 9000/' \
	-e 's/"UseFastCGI": false/"UseFastCGI": true/' \
	-e 's/"Username": ".*",//' \
	-e 's#"DocumentRoot": "html"#"DocumentRoot": "/srv/gochan"#' \
	-e 's#"TemplateDir": "templates"#"TemplateDir": "/usr/share/gochan/templates"#' \
	-e 's#"LogDir": "log"#"LogDir": "/var/log/gochan"#' \
	-e 's/"Verbosity": 0/"Verbosity": 1/' \
	-e "s/\"DBtype\".*/\"DBtype\": \"${DBTYPE}\",/" \
	-e "s/\"DBhost\".*/\"DBhost\": \"tcp(${DATABASE_HOST}:${DATABASE_PORT})\",/" \
	-e "s/\"DBname\".*/\"DBname\": \"${DATABASE_NAME}\",/" \
	-e "s/\"DBusername\".*/\"DBusername\": \"${DATABASE_USER}\",/" \
	-e "s/\"DBpassword\".*/\"DBpassword\": \"${DATABASE_PASSWORD}\",/" \
	&& mkdir -p /etc/ssl/private \
	&& openssl req -x509 -nodes -days 7305 -newkey rsa:2048 -keyout /etc/ssl/private/nginx-selfsigned.key -out /etc/ssl/certs/nginx-selfsigned.crt -subj "/CN=127.0.0.1" \
	&& ./build.py \
	&& ./build.py install \
	&& nginx \
	&& echo "pinging db" \
	&& docker/wait-for.sh $DATABASE_HOST:$DATABASE_PORT -t 30 \
	&& /opt/gochan/gochan -rebuild all \
	&& /opt/gochan/gochan
