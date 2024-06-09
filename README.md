# Proposed project: Internet Download Manager
- FE-facing Golang HTTP server using grpc
- Message consumer passively handling download tasks
- Cronjob running periodcally to actively handling download tasks
- Downloaded file can be configured to saved locally or uploaded to a self-hosted S3 server

# Tech stack
- Go HTTP server: Standard library net/http plus grpc
- Database: MySQL, but we should have a generic implementation for any database engine
- Cache: Redis, but we should have a generic implementation for any cache engine
- Message queue: Kafka, but we should have a generic implementation for any message queue
- Block storage: We should have a generic implementation shared between S3 and Local File System
