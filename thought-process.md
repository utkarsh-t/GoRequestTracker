High-Level Overview of the Design and Implementation Approach
The goal of this application is to handle 10K requests per second efficiently, ensure deduplication across multiple instances, and provide extensions like sending unique request counts via HTTP requests or Kafka. Here's the breakdown of the approach taken:

1. Concurrency and High-Performance Request Handling
Go’s HTTP Package: The net/http package in Go is used to build a REST API because it is highly efficient, lightweight, and inherently supports concurrency by spawning a new goroutine for each incoming request. This allows the service to scale and handle a large number of requests per second.

sync.Map for In-Memory Deduplication: While handling the uniqueness of ids, I used sync.Map for thread-safe in-memory storage of requests within a single instance. However, to ensure cross-instance deduplication, this was extended with Redis.

2. Cross-Instance Deduplication
Redis as a Central Store: For deduplication across instances behind a load balancer, I integrated Redis as a centralized data store. Redis provides fast and atomic operations for checking and adding unique ids. The SAdd operation ensures that only unique IDs are added, while the SCard operation efficiently counts the total number of unique ids.
Why Redis?: Redis is an ideal choice for this because it is an in-memory data store, making it very fast. Redis SETs also inherently guarantee uniqueness, making the deduplication logic straightforward.
3. Handling of HTTP Requests
GET Request Handling: The /api/verve/accept endpoint accepts an integer id as a query parameter and optionally an endpoint. The service checks the uniqueness of the id, logs whether it is a duplicate, and returns either "ok" or "failed" depending on any errors.

Optional Endpoint: When the endpoint parameter is provided, the service fires either a GET or a POST request (as per the extension). The request includes the count of unique requests received in the current minute as a query parameter (for GET) or in the JSON body (for POST).

4. Logging the Count of Unique Requests
Kafka Integration for Distributed Logging: Instead of logging the unique request count to a local file, I integrated Kafka for distributed logging. This ensures that the logs can be collected and processed in a scalable, fault-tolerant manner. Each minute, the unique request count is sent to a Kafka topic (unique-requests).

Kafka Producer: The Kafka producer is initialized at the start of the application and sends log messages with the current timestamp as the key and the unique request count as the value. Kafka’s distributed nature ensures that logs are handled reliably even as the system scales.

5. Extension 1: POST Request Handling
Instead of just sending GET requests to the optional endpoint, I added support for POST requests as an extension. The POST request sends a JSON payload with the structure:
json
Copy code
{
  "unique_count": <count>
}
This allows flexibility for more complex interactions with external systems where a POST request with a structured payload is required.
6. Extension 2: Cross-Instance Deduplication (Behind a Load Balancer)
This task’s main challenge was to ensure deduplication works even when multiple instances are deployed behind a load balancer. By using Redis, all instances share a single, fast, and atomic store for tracking unique ids.
Each instance communicates with Redis to add new IDs and checks if the current id has already been seen, ensuring that deduplication works across all instances.
7. Extension 3: Distributed Streaming
The requirement to log the count of unique requests to a distributed system was implemented using Kafka. Instead of writing to a log file, the count of unique IDs is sent to a Kafka topic. This allows for distributed processing of log data and more advanced analysis downstream, like monitoring or further aggregation.
Design Considerations:
Scalability:

The use of Redis ensures that the application can scale horizontally, with multiple instances tracking deduplication centrally.
Kafka as a distributed logging system supports scalability by handling high-throughput log events efficiently.
Performance:

Go’s inherent concurrency, combined with Redis for fast deduplication and Kafka for asynchronous log processing, ensures the application meets the requirement to handle 10K requests per second.
Fault Tolerance:

Redis: If Redis is temporarily unavailable, the system could fallback to in-memory deduplication using sync.Map, although it would lose cross-instance consistency.
Kafka: Kafka ensures logs are distributed reliably, but it also has built-in resilience for temporary failures.
Maintainability and Extensibility:

By decoupling the deduplication logic (Redis) and logging logic (Kafka), each component can be scaled, modified, or replaced independently if needed.
The POST request extension demonstrates how easy it is to add new features, such as alternative ways to handle external API requests.
Future Enhancements:
Caching: Introduce caching for Redis to reduce latency in checking unique IDs.
Monitoring: Add Prometheus for real-time metrics to monitor performance, including request rates and unique counts.
Retry Mechanism: Implement retries for Redis and Kafka in case of temporary failures.
Conclusion:
This design ensures high performance, scalability, and flexibility while maintaining simplicity and adhering to the requirements. The use of Redis and Kafka makes the system ready for production-level workloads and scalable distributed environments.