package service

import (
	"strings"
	"testing"
	"time"
)

func TestPresignMatchesAWSReferenceSignature(t *testing.T) {
	s3, err := NewS3(S3Config{
		Endpoint:  "https://s3.amazonaws.com",
		Region:    "us-east-1",
		Bucket:    "examplebucket",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		PathStyle: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	at, _ := time.Parse("20060102T150405Z", "20130524T000000Z")
	got := s3.Presign("GET", "test.txt", 86400*time.Second, at)
	want := "X-Amz-Signature=aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404"
	if !strings.HasSuffix(got, want) {
		t.Fatalf("signature mismatch:\n got %s\nwant suffix %s", got, want)
	}
}
