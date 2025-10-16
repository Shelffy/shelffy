package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

type FileStorage interface {
	Upload(ctx context.Context, path string, contentLength int64, bookContent io.Reader) error
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	BatchDelete(ctx context.Context, paths ...string) ([]NotDeleted, error)
}

var (
	ErrObjectNotFound = errors.New("object not found")
)

type s3storage struct {
	bucketName string
	s3client   *s3.Client
}

func NewS3Storage(bucket string, client *s3.Client) FileStorage {
	return s3storage{
		bucketName: bucket,
		s3client:   client,
	}
}

func (s s3storage) Upload(ctx context.Context, path string, contentLength int64, bookContent io.Reader) error {
	_, err := s.s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(path),
		ContentLength: aws.Int64(contentLength),
		Body:          bookContent,
	})
	return err
}

func (s s3storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	out, err := s.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) && aerr.Code() == "NotFound" {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return out.Body, nil
}

func (s s3storage) Delete(ctx context.Context, path string) error {
	_, err := s.s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return ErrObjectNotFound
		}
		return err
	}
	err = s3.NewObjectNotExistsWaiter(s.s3client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucketName), Key: aws.String(path)}, time.Minute)
	if err != nil {
		return err
	}
	return nil
}

type NotDeleted struct {
	Path  string
	Cause error
}

func (s s3storage) BatchDelete(ctx context.Context, paths ...string) ([]NotDeleted, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	objects := make([]types.ObjectIdentifier, len(paths))
	for i, path := range paths {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(path),
		}
	}
	out, err := s.s3client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})
	if out != nil && len(out.Errors) > 0 {
		notDeleted := make([]NotDeleted, 0, len(out.Errors))
		for _, e := range out.Errors {
			notDeleted = append(notDeleted, NotDeleted{
				Path:  aws.ToString(e.Key),
				Cause: fmt.Errorf("%s", aws.ToString(e.Message)),
			})
		}
		return notDeleted, err
	}
	return nil, err
}
