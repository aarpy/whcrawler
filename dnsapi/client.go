package main

import (
	"fmt"

	"gopkg.in/redis.v3"
)

// redis client
var client *redis.Client

func main2() {
	// new client
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// close client
	if client != nil {
		defer client.Close()
	}

	// Test 1
	poingNewClient()

	// Test 2
	testClient()
}

func poingNewClient() {

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
}

func testClient() {
	err := client.Set("key", "goal", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exists")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exists
}
