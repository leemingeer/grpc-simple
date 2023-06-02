# 生成.key  私钥文件
openssl genrsa -out ca.key 2048
# 生成.csr 证书签名请求文件
openssl req -new -key ca.key -out ca.csr  -subj "/C=GB/L=China/O=lixd/CN=www.ming.com"
# 自签名生成.crt 证书文件
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt  -subj "/C=GB/L=China/O=lixd/CN=www.ming.com"

openssl genrsa -out server.key 2048

# 生成.key  私钥文件
openssl genrsa -out client.key 2048

# 生成.csr 证书签名请求文件
openssl req -new -key client.key -out client.csr \
	-subj "/C=GB/L=China/O=ming/CN=www.ming.com" \
	-reqexts SAN \
	-config <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.ming.com,DNS:*.test.com"))

# 签名生成.crt 证书文件
openssl x509 -req -days 3650 \
   -in client.csr -out client.crt \
   -CA ca.crt -CAkey ca.key -CAcreateserial \
   -extensions SAN \
   -extfile <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.ming.com,DNS:*.test.com"))


# 生成.csr 证书签名请求文件
openssl req -new -key server.key -out server.csr \
	-subj "/C=GB/L=China/O=ming/CN=www.ming.com" \
	-reqexts SAN \
	-config <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.ming.com,DNS:*.test.com"))

# 签名生成.crt 证书文件
openssl x509 -req -days 3650 \
   -in server.csr -out server.crt \
   -CA ca.crt -CAkey ca.key -CAcreateserial \
   -extensions SAN \
   -extfile <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.ming.com,DNS:*.text.com"))
