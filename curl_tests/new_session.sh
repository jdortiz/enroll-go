curl http://127.0.0.1:8080/session -v \
     -H 'Content-Type: application/json' \
     -H 'Accept: application/json' \
     -d '{"username":"alice", "password": "secret123"}'
