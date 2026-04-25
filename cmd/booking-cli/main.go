package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type client struct {
	baseURL string
	token   string
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		return
	}

	baseURL := getenv("BOOKING_API", "http://localhost:8080")
	token := os.Getenv("BOOKING_TOKEN")
	c := client{baseURL: baseURL, token: token}

	switch os.Args[1] {
	case "register":
		register(c, os.Args[2:])
	case "login":
		login(c, os.Args[2:])
	case "rooms":
		rooms(c, os.Args[2:])
	case "create-room":
		createRoom(c, os.Args[2:])
	case "book":
		book(c, os.Args[2:])
	case "my-bookings":
		get(c, "/bookings/me", os.Args[2:])
	case "bookings":
		get(c, "/bookings", os.Args[2:])
	case "cancel":
		cancel(c, os.Args[2:])
	case "audit":
		get(c, "/audit-logs", os.Args[2:])
	default:
		usage()
	}
}

func register(c client, args []string) {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	username := fs.String("username", "", "username")
	password := fs.String("password", "", "password")
	role := fs.String("role", "user", "role")
	adminCode := fs.String("admin-code", "", "admin code")
	fs.Parse(args)

	body := map[string]interface{}{
		"username":   *username,
		"password":   *password,
		"role":       *role,
		"admin_code": *adminCode,
	}
	c.request(http.MethodPost, "/register", "", body)
}

func login(c client, args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	username := fs.String("username", "", "username")
	password := fs.String("password", "", "password")
	fs.Parse(args)

	body := map[string]interface{}{"username": *username, "password": *password}
	c.request(http.MethodPost, "/login", "", body)
}

func rooms(c client, args []string) {
	fs := flag.NewFlagSet("rooms", flag.ExitOnError)
	token := fs.String("token", c.token, "token")
	start := fs.String("start", "", "start time")
	end := fs.String("end", "", "end time")
	capacity := fs.Int("capacity", 0, "minimum capacity")
	fs.Parse(args)

	values := url.Values{}
	if *start != "" {
		values.Set("start_time", normalizeTime(*start))
	}
	if *end != "" {
		values.Set("end_time", normalizeTime(*end))
	}
	if *capacity > 0 {
		values.Set("capacity", strconv.Itoa(*capacity))
	}

	path := "/rooms"
	if values.Encode() != "" {
		path += "?" + values.Encode()
	}
	c.request(http.MethodGet, path, *token, nil)
}

func createRoom(c client, args []string) {
	fs := flag.NewFlagSet("create-room", flag.ExitOnError)
	token := fs.String("token", c.token, "token")
	name := fs.String("name", "", "room name")
	capacity := fs.Int("capacity", 0, "capacity")
	location := fs.String("location", "", "location")
	fs.Parse(args)

	body := map[string]interface{}{
		"name":     *name,
		"capacity": *capacity,
		"location": *location,
	}
	c.request(http.MethodPost, "/rooms", *token, body)
}

func book(c client, args []string) {
	fs := flag.NewFlagSet("book", flag.ExitOnError)
	token := fs.String("token", c.token, "token")
	roomID := fs.Int("room", 0, "room id")
	start := fs.String("start", "", "start time")
	end := fs.String("end", "", "end time")
	fs.Parse(args)

	body := map[string]interface{}{
		"room_id":    *roomID,
		"start_time": normalizeTime(*start),
		"end_time":   normalizeTime(*end),
	}
	c.request(http.MethodPost, "/bookings", *token, body)
}

func cancel(c client, args []string) {
	fs := flag.NewFlagSet("cancel", flag.ExitOnError)
	token := fs.String("token", c.token, "token")
	id := fs.Int("id", 0, "booking id")
	fs.Parse(args)

	c.request(http.MethodDelete, "/bookings/"+strconv.Itoa(*id), *token, nil)
}

func get(c client, path string, args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	token := fs.String("token", c.token, "token")
	fs.Parse(args)
	c.request(http.MethodGet, path, *token, nil)
}

func (c client) request(method string, path string, token string, body interface{}) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			log.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode >= 400 {
		log.Fatalf("%s: %s", res.Status, string(data))
	}

	var output interface{}
	if err := json.Unmarshal(data, &output); err != nil {
		fmt.Println(string(data))
		return
	}
	pretty, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(pretty))
}

func normalizeTime(value string) string {
	if value == "" {
		return value
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format(time.RFC3339)
		}
	}
	return value
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func usage() {
	fmt.Println("meeting room booking cli")
	fmt.Println("commands:")
	fmt.Println("  register -username name -password pass")
	fmt.Println("  login -username name -password pass")
	fmt.Println("  rooms -token token")
	fmt.Println("  create-room -token token -name room -capacity 8")
	fmt.Println("  book -token token -room 1 -start \"2026-05-01 10:00\" -end \"2026-05-01 11:00\"")
	fmt.Println("  my-bookings -token token")
	fmt.Println("  bookings -token token")
	fmt.Println("  cancel -token token -id 1")
	fmt.Println("  audit -token token")
}
