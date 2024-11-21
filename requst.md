**Post Request**

`
encoded_value=$(echo -n "Sanchit Sharma" | base64)
curl -X POST localhost:8080/produce \
     -H "Content-Type: application/json" \
     -d '{"record": {"value": "'"$encoded_value"'"}}'
`

**Get Request**
`curl -X GET localhost:8080/produce -d '{"offset": 2}'`