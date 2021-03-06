nginx-build \
 -idempotent \
 -d work \
 -openresty \
 -openrestyversion=1.13.6.2 \
 -zlib \
 -zlibversion=1.2.11 \
 -pcre \
 -pcreversion=8.42 \
 -verbose \
 -clear \
 --prefix=/etc/nginx \
 --sbin-path=/usr/sbin/nginx \
 --conf-path=/etc/nginx/nginx.conf \
 --error-log-path=/var/log/nginx/error.log \
 --http-log-path=/var/log/nginx/access.log \
 --pid-path=/var/run/nginx.pid \
 --lock-path=/var/run/nginx.lock \
 --user=nginx \
 --group=nginx \
 --with-http_realip_module \
 --with-http_stub_status_module \
 --with-http_gzip_static_module \
 --with-http_gunzip_module \
 --with-file-aio \
 --with-cc-opt='-O2 -g -pipe -Wp,-D_FORTIFY_SOURCE=2 -fexceptions -fstack-protector --param=ssp-buffer-size=4 -m64 -mtune=generic' \
 --with-pcre-jit

