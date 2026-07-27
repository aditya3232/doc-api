## 1. build aplikasi
- docker build -t doc-api-service .

## 2. start aplikasi
- cd deployment
- docker compose up -d --build

## 3. login HA Proxy
- http://localhost:8404/stats
- username: admin
- password: admin123