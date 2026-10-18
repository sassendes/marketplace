package service

import "testing"

func TestTokenBucketBurstThenBlock(t *testing.T) {
	rl := NewRateLimiter(1, 3)
	for i := 0; i < 3; i++ {
		if ok, _ := rl.Allow("a"); !ok {
			t.Fatalf("request %d should pass", i)
		}
	}
	if ok, wait := rl.Allow("a"); ok || wait <= 0 {
		t.Fatal("4th request should be limited with a retry hint")
	}
	if ok, _ := rl.Allow("b"); !ok {
		t.Fatal("other clients must have their own bucket")
	}
}
