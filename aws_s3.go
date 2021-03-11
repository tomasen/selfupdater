package selfupdater

import (
	"crypto/md5"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"hash"
	"os"
)

type S3UpdateProvider struct {
	BucketName     string
	ExecutablePath string
	ChecksumPath   string
	S3Downloader   *s3manager.Downloader
}

func NewS3UpdateProvider(bucket string, region string, executable string, checksum string) *S3UpdateProvider {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}))
	downloader := s3manager.NewDownloader(sess)

	return &S3UpdateProvider{BucketName: bucket,
		ExecutablePath: executable,
		ChecksumPath:   checksum,
		S3Downloader:   downloader,
	}
}

func (p *S3UpdateProvider) DownloadTo(file *os.File) error {
	_, err := p.S3Downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(p.BucketName),
			Key:    aws.String(p.ExecutablePath),
		})
	if err != nil {
		return errors.WithMessage(err, "failed to obtain remote executable")
	}
	return nil
}

func (p *S3UpdateProvider) RemoteChecksum() (string, error) {

	buff := aws.NewWriteAtBuffer([]byte{})
	numBytes, err := p.S3Downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(p.BucketName),
			Key:    aws.String(p.ChecksumPath),
		})

	if err != nil {
		return "", errors.WithMessage(err, "failed to obtain remote checksum")
	}
	return string(buff.Bytes()[:numBytes]), nil
}

func (p *S3UpdateProvider) Hash() hash.Hash {
	return md5.New()
}
