package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"golang.org/x/net/context"
)

// Redis client
var redisClient *redis.Client
var ctx = context.Background()

// Get environment variables
var redisHost = os.Getenv("REDIS_URL")
var kafkaBroker = os.Getenv("KAFKA_BROKER")

// Kafka writer
var kafkaWriter *kafka.Writer

func main() {
	// Initialize Redis
	initRedis()

	// Initialize Kafka
	initKafka()

	// Logger
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	go logUniqueCountEveryMinute(logger)

	http.HandleFunc("/api/verve/accept", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Initialize Redis connection
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost, // Redis address
	})
}

// Initialize Kafka producer
func initKafka() {
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    "unique-requests",
		Balancer: &kafka.LeastBytes{},
	}
}

// Handle incoming requests
func handleRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	endpoint := r.URL.Query().Get("endpoint")
	method := r.URL.Query().Get("method") // New parameter for method

	if id == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	// Track unique id in Redis
	if isNewID(id) {
		log.Printf("New unique request: %s\n", id)
	} else {
		log.Printf("Duplicate request: %s\n", id)
	}

	// Send GET/POST request to endpoint if provided
	if endpoint != "" {
		if method == "POST" {
			sendPostRequestToEndpoint(endpoint)
		} else {
			sendRequestToEndpoint(endpoint)
		}
	}

	fmt.Fprintf(w, "ok")
}

// Check if the id is unique across instances using Redis SET
func isNewID(id string) bool {
	result, err := redisClient.SAdd(ctx, "unique_ids", id).Result()
	if err != nil {
		log.Printf("Error adding id to Redis: %v", err)
		return false
	}
	// result == 1 means it's a new unique id
	return result == 1
}

// Log unique request count every minute to Kafka
func logUniqueCountEveryMinute(logger *log.Logger) {
	for {
		time.Sleep(1 * time.Minute)

		// Get unique count from Redis
		count, err := redisClient.SCard(ctx, "unique_ids").Result()
		if err != nil {
			logger.Printf("Error getting unique request count: %v", err)
			continue
		}

		// Log unique request count to Kafka
		sendCountToKafka(int(count))

		// Reset Redis set for the next minute
		redisClient.Del(ctx, "unique_ids")
	}
}

// Send unique request count to Kafka
func sendCountToKafka(count int) {
	msg := kafka.Message{
		Key:   []byte(strconv.Itoa(int(time.Now().Unix()))),
		Value: []byte(fmt.Sprintf("Unique requests in last minute: %d", count)),
	}
	err := kafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Printf("Error sending count to Kafka: %v", err)
	} else {
		log.Printf("Sent unique request count to Kafka: %d\n", count)
	}
}

// Send HTTP request to the provided endpoint with unique request count
func sendRequestToEndpoint(endpoint string) {
	count, err := redisClient.SCard(ctx, "unique_ids").Result()
	if err != nil {
		log.Printf("Error getting unique request count from Redis: %v", err)
		return
	}

	url := fmt.Sprintf("%s?count=%d", endpoint, count)

	// Perform GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error sending request to %s: %v", endpoint, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Sent request to %s, Status: %d\n", endpoint, resp.StatusCode)
}

// Extension 1: Send POST request instead of GET
func sendPostRequestToEndpoint(endpoint string) {
	count, err := redisClient.SCard(ctx, "unique_ids").Result()
	if err != nil {
		log.Printf("Error getting unique request count from Redis: %v", err)
		return
	}

	data := map[string]int{
		"unique_count": int(count),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Error sending POST request to %s: %v", endpoint, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Sent POST request to %s, Status: %d\n", endpoint, resp.StatusCode)
}
