package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestGroupingInvisibleToS3(t *testing.T) {
	ts := setupTestServer(t)
	ctx := context.Background()
	bucket := "ecom-products"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	before, err := ts.Client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}

	// set a dashboard-only group straight on the store
	if err := ts.Store.UpdateBucketGroup(bucket, "ecom"); err != nil {
		t.Fatalf("UpdateBucketGroup: %v", err)
	}

	after, err := ts.Client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets after: %v", err)
	}
	if len(after.Buckets) != len(before.Buckets) {
		t.Fatalf("S3 bucket count changed after grouping: %d -> %d", len(before.Buckets), len(after.Buckets))
	}
	found := false
	for _, b := range after.Buckets {
		if aws.ToString(b.Name) == bucket {
			found = true
		}
	}
	if !found {
		t.Fatal("bucket missing from S3 ListBuckets after grouping")
	}
	if _, err := ts.Client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("HeadBucket after grouping: %v", err)
	}
}

func TestPinningInvisibleToS3(t *testing.T) {
	ts := setupTestServer(t)
	ctx := context.Background()
	bucket := "saas-avatars"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	before, err := ts.Client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}

	// set a dashboard-only pin straight on the store
	if err := ts.Store.UpdateBucketPinned(bucket, true); err != nil {
		t.Fatalf("UpdateBucketPinned: %v", err)
	}

	after, err := ts.Client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets after: %v", err)
	}
	if len(after.Buckets) != len(before.Buckets) {
		t.Fatalf("S3 bucket count changed after pinning: %d -> %d", len(before.Buckets), len(after.Buckets))
	}
	found := false
	for _, b := range after.Buckets {
		if aws.ToString(b.Name) == bucket {
			found = true
		}
	}
	if !found {
		t.Fatal("bucket missing from S3 ListBuckets after pinning")
	}
	if _, err := ts.Client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("HeadBucket after pinning: %v", err)
	}
}
