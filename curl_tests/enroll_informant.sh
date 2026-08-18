curl http://127.0.0.1:8080/informant -v \
     -H 'Content-Type: application/json' \
     -H 'Accept: application/json' \
     -H "Authorization: Bearer $SESSION_JWT" \
     -d '{"alias": "bob", "phone": "990123400", "country": "GB", "area": "London" }'
