set -e
rm -rf ./act/certs && mkdir -p ./act/certs
cd ./act/certs/
security find-certificate -a -c "Gateway CA - Cloudflare Managed" -p | \
awk '/-----BEGIN CERTIFICATE-----/{file="warp" ++i ".pem"} {print > file}'
for pem in warp*.pem; do
  openssl x509 -in "$pem" -inform PEM -out "${pem%.pem}.crt"
  rm "$pem"
done
docker build -t act-warp ../