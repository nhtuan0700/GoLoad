curl -X POST http://localhost:8080/go_load.GoLoadService/CreateAccount \
  -H "Content-Type: application/json" \
  -d '{"accountName": "user11", "password": "password"}'
