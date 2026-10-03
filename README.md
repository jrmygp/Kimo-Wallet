<!-- Temporary command reminder -->
buf lint
buf generate --template buf.gen.user-service.yaml
buf generate --template buf.gen.api-gateway.yaml

generate protobuf service folder in target service:
buf generate --template buf.gen.api-gateway.yaml --path packages/contracts/transaction/v1/transaction.proto

IMPORTANT : create the proto contract inside packages folder!

How to "export-import" service's functions and methods:
1. Run the command above
2. Add <NAME>_SERVICE_GRPC_ADDR in config var
3. Open grpc client connection in your service 