curl http://127.0.0.1:8080/session/verification -v \
     -H 'Content-Type: application/json' \
     -H 'Accept: application/json' \
     -d "{\"request_id\": \"${REQUEST_ID}\", \"username\": \"alice\", \"code\": \"$CODE\"}"
