package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/crisanp13/shop/src/types"
)

var baseUrl url.URL

func TestMain(m *testing.M) {
	port := ":" + getEnv("SHOP_PORT")
	url, err := url.Parse("http://localhost" + port)
	baseUrl = *url
	if err != nil {
		fmt.Println("failed to parse base url:", err)
	}
	code := m.Run()
	os.Exit(code)
}

func getEnv(key string) string {
	env := map[string]string{
		"SHOP_PORT": "8080",
	}
	return env[key]
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

func getEndpoint(path string) string {
	return baseUrl.JoinPath(path).String()
}

func TestRegister(t *testing.T) {
	var buf bytes.Buffer
	registerReq := types.RegisterReq{
		Name:     "Jim Jomson",
		Email:    "jimjimson@gmail.com",
		Password: "Pass1!",
	}
	err := json.NewEncoder(&buf).Encode(registerReq)
	if err != nil {
		t.Fatal("failed to encode register req,", err)
	}
	req, err := http.NewRequest(http.MethodPost,
		getEndpoint("user/register"),
		&buf)
	if err != nil {
		t.Fatal("failed to create register req:", err)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("failed to send register req:", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("failed to read register resp,", err)
		}
		t.Fatalf("did not receive OK on register, received: %d, with: %s",
			resp.StatusCode,
			string(body))
	}
	var registerResp types.RegiesterResp
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	if err != nil {
		t.Fatal("failed to decode register res")
	}
	if registerResp.Id == "" {
		t.Fatal("id is empty")
	}
	resp.Body.Close()
	loginReq := types.LoginReq{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}
	buf.Reset()
	err = json.NewEncoder(&buf).Encode(loginReq)
	if err != nil {
		t.Fatal("failed ot encode login req")
	}
	req, err = http.NewRequest(http.MethodPost,
		getEndpoint("user/login"),
		&buf)
	if err != nil {
		t.Fatal("failed to create login req:", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal("failed to send login req,", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("failed to read login resp,", err)
		}
		t.Fatalf("did not receive OK from loign, received: %d, with %s",
			resp.StatusCode,
			string(body))
	}
	var loginResp types.LoginResp
	if err = json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatal("failed to decode login resp")
	}
	resp.Body.Close()
	if len(loginResp.Token) == 0 ||
		!strings.HasPrefix(loginResp.Token, "Bearer:") {
		t.Fatal("invalid auth header:", loginResp.Token)
	}
}

func TestAuthWithUserDetails(t *testing.T) {
	loginReq := types.LoginReq{Email: "johnjohnson@gmail.com", Password: "Pass1!"}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(loginReq)
	if err != nil {
		t.Fatal("error encoding login req", err)
	}
	req, err := http.NewRequest(http.MethodPost,
		getEndpoint("user/login"),
		&buf)
	if err != nil {
		t.Fatal("failed to create login req", err)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatal("failed login req", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	var loginResp types.LoginResp
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		t.Fatal("failed to decode login resp", err)
	}
	req, err = http.NewRequest(http.MethodGet,
		getEndpoint("user/details/"+loginResp.Id),
		nil)
	if err != nil {
		t.Fatal("failed to make detail req", err)
	}
	req.Header.Add("Authorization", loginResp.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal("failed details req", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("failed to read login resp,", err)
		}
		t.Fatal("failed details req,", resp.StatusCode, string(b))
	}
	var user types.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatal("failed to decode user,", err)
	}
	if user.Id != loginResp.Id {
		t.Errorf("user id differs, %s != %s", user.Id, loginResp.Id)
	}
}
