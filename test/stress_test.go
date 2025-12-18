// filepath: /Users/shiling/project/seckill-system/test/stress_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	t.Log("🧪 开始测试限流功能")

	token := login("testuser", "123456")
	if token == "" {
		t.Fatal("登录失败，未获取到 token")
	}

	for i := 1; i <= 3; i++ {
		resp := seckill(token)
		t.Logf("第%d次: %s", i, resp)
	}

	t.Log("\n⏰ 等待2秒后再试...")
	time.Sleep(2 * time.Second)

	resp := seckill(token)
	t.Logf("2秒后: %s", resp)
}

func login(username, password string) string {
	data := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(data)
	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if token, ok := result["token"].(string); ok {
		return token
	}
	return ""
}

func seckill(token string) string {
	req, _ := http.NewRequest("POST", "http://localhost:8080/user/seckill/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("❌ 请求失败: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if msg, ok := result["message"]; ok {
		return fmt.Sprintf("✅ %v", msg)
	}
	if errMsg, ok := result["error"]; ok {
		return fmt.Sprintf("❌ %v", errMsg)
	}
	return "❌ 未知响应"
}
