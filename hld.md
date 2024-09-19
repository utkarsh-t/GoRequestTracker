

1. Technology Choice (Go)
Go is preferred for high-concurrency tasks due to its lightweight goroutines and low-latency processing, which are essential for handling 10K requests per second.
2. Application Structure
GET Endpoint: /api/verve/accept with an integer id as a required parameter and an optional HTTP endpoint.
Logger: A standard logger will log the count of unique requests every minute.
Concurrency: Use goroutines for handling multiple requests and mutex for ensuring thread-safe operations.
3. Unique Request Tracking
Use a sync.Map to track unique IDs for each minute. This ensures efficient read/write access in a concurrent environment.
At the end of every minute, calculate the number of unique requests and reset the map for the next minute.
4. HTTP Requests
If an endpoint is provided, fire an HTTP GET request with the unique count as a query parameter. Log the status code using the logger.
5. Extensions
Extension 1: Modify the logic to use an HTTP POST request instead of GET. The body of the POST request can contain the unique request count in JSON format.
Extension 2: Implement deduplication across multiple instances by using a distributed cache such as Redis. Each instance can check Redis for existing IDs and store new ones to ensure deduplication across instances.
Extension 3: Instead of logging, send the count of unique requests to a distributed streaming platform such as Kafka.
6. Scalability
Use Docker to containerize the application for easy deployment and scaling.
To handle load balancing, ensure that the deduplication strategy works effectively by sharing data between instances.
7. File Structure
Main Application: The main server will handle incoming requests, track unique IDs, and manage concurrent logging.
thought-process.md: Describes the rationale for design choices, concurrency handling, and scalability considerations.
Tools :

Go’s HTTP Package for request handling.
sync.Map and goroutines for concurrency.
Redis for cross-instance deduplication.
Kafka for distributed logging.


This high-level approach ensures the application meets the performance and extension requirements effectively.