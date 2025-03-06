package test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/crisanp13/shop/src/api"
	"github.com/crisanp13/shop/src/types"
)

var baseUrl url.URL
var port string

func TestMain(m *testing.M) {
	flag.StringVar(&port, "port", "8080", "port to run server on")
	flag.Parse()
	port = ":" + port
	url, err := url.Parse("http://localhost" + port)
	baseUrl = *url
	if err != nil {
		fmt.Println("failed to parse base url:", err)
	}
	os.Exit(m.Run())
}

func TestHealthCheck(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	logger := createLog()
	go api.Run(logger, port, ctx)
	err := waitForReady(ctx, time.Second, getEndpoint("health"))
	if err != nil {
		t.Error("wait for endpoint failed,", err.Error())
	}
	t.Log("running")
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go api.Run(log.Default(), port, ctx)
	err := waitForReady(ctx, time.Second, getEndpoint("health"))
	if err != nil {
		t.Fatal("wait for endpoint failed,", err.Error())
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(
		types.RegisterReq{
			Name:     "Jim Jomson",
			Email:    "jimjimson@gmail.com",
			Password: "Pass1!",
		})
	if err != nil {
		t.Fatal("failed to encode request,", err)
	}
	req, err := http.NewRequest(http.MethodPost,
		getEndpoint("user/register"),
		&buf)
	if err != nil {
		t.Fatal("failed to create request:", err)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("failed to send request:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("failed to read response body")
		}
		t.Fatalf("did not receive OK, received: %d, with: %s",
			resp.StatusCode,
			string(body))
	}
	var registerResp types.RegiesterResp
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	if err != nil {
		t.Fatal("failed to decode response")
	}
	if registerResp.Id == "" {
		t.Fatal("id is empty")
	}
	t.Logf("Id: %s", registerResp.Id)
}

func createLog() *log.Logger {
	var buf bytes.Buffer
	return log.New(&buf, "", log.LstdFlags)
}

func getEndpoint(path string) string {
	return baseUrl.JoinPath(path).String()
}

func waitForReady(
	ctx context.Context,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request, %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("error making request, ", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reaced while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
