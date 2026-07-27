## 1. build aplikasi
- docker build -t doc-api-service .

## 2. start aplikasi
- cd deployment
- docker compose up -d --build

## 3. login HA Proxy
- http://localhost:8404/stats
- username: admin
- password: admin123

## 4. buat ssl 
- cd deployment/haproxy/certs
```bash
openssl req \
    -x509 \
    -newkey rsa:4096 \
    -sha256 \
    -days 365 \
    -nodes \
    -keyout server.key \
    -out server.crt \
    -subj "/CN=localhost"

cat server.crt server.key > server.pem
```
- kalau berhasil hasilnya akan seperti ini:
```bash
deployment/
└── haproxy/
    ├── certs/
    │   ├── server.crt   # Public Certificate
    │   ├── server.key   # Private Key
    │   └── server.pem   # Gabungan crt + key (dipakai HAProxy)
    └── haproxy.cfg
```