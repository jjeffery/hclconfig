package amzn

import (
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jjeffery/errors"
)

// Get the contents of an S3 bucket. The caller is responsible for
// closing the body.
func Get(bucket, key string) (etag string, modified time.Time, body io.ReadCloser, err error) {
	s3svc := s3.New(AWSSession())
	output, err := s3svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		err = errors.Wrap(err, "cannot download from S3").With(
			"bucket", bucket,
			"key", key,
		)
		return etag, modified, body, err
	}
	if output.ETag != nil {
		etag = *output.ETag
	}
	if output.LastModified != nil {
		modified = *output.LastModified
	}
	body = output.Body
	return etag, modified, body, nil
}

// Head the contents of an S3 bucket.
func Head(bucket, key string) (etag string, modified time.Time, err error) {
	s3svc := s3.New(AWSSession())
	output, err := s3svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		err = errors.Wrap(err, "cannot download from S3").With(
			"bucket", bucket,
			"key", key,
		)
		return etag, modified, err
	}
	if output.ETag != nil {
		etag = *output.ETag
	}
	if output.LastModified != nil {
		modified = *output.LastModified
	}
	return etag, modified, nil
}

// HasChanged determines whether the S3 object has changed.
func HasChanged(bucket, key string, etag string) (changed bool, err error) {
	// We don't bother with last modified because we know S3 always
	// returns an ETag and passing both IfNoneMatch and IfModifiedSince
	// only complicates things as per RFC 7232.
	s3svc := s3.New(AWSSession())
	_, err = s3svc.HeadObject(&s3.HeadObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		IfNoneMatch: aws.String(etag),
	})
	type statusCoder interface {
		StatusCode() int
	}
	if err != nil {
		if statusCode, ok := err.(statusCoder); ok {
			if statusCode.StatusCode() == http.StatusNotModified {
				// not modified
				return false, nil
			}
		}
		return false, errors.Wrap(err, "cannot HEAD S3 object").With(
			"bucket", bucket,
			"key", key,
		)
	}
	// object has changed
	return true, nil
}
