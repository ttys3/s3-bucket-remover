package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/schollz/progressbar/v3"
)

var appVer = "1.0.0"

// Site is a option for backing up data to S3
type Site struct {
	Bucket       string `yaml:"bucket"`
	BucketPath   string `yaml:"bucket_path"`
	BucketRegion string `yaml:"bucket_region"`
	Endpoint     string `yaml:"endpoint"`
	StorageClass string `yaml:"storage_class"`
	AccessKey    string `yaml:"access_key"`
	Secret       string `yaml:"secret"`
}

func main() {
	var bucket string
	var bucketPath string
	var endpoint string
	var region string
	var logLevel string
	var accessKey string
	var secret string
	var showVerOnly bool

	// Read command line args
	flag.StringVar(&bucket, "b", "", "bucket to remove")
	flag.StringVar(&bucketPath, "p", "/", "bucket path prefix to remove")
	flag.StringVar(&endpoint, "e", "", "endpoint")
	flag.StringVar(&region, "r", "us-east-1", "region")
	flag.StringVar(&accessKey, "k", "", "access Key")
	flag.StringVar(&secret, "s", "", "secret")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.BoolVar(&showVerOnly, "v", false, "show app version")
	flag.Parse()

	fmt.Printf(" ==== s3 bucket remover %s ====\n", appVer)

	if showVerOnly {
		os.Exit(0)
	}
	// init logger
	initLogger(logLevel)

	site := Site{
		Bucket:       bucket,
		BucketPath:   bucketPath,
		BucketRegion: region,
		Endpoint:     endpoint,
		AccessKey:    accessKey,
		Secret:       secret,
	}

	// Remove leading slash from the BucketPath
	site.BucketPath = strings.TrimLeft(site.BucketPath, "/")
	logger.Debugf("site %#v", site)

	cred := credentials.NewStaticCredentials(site.AccessKey, site.Secret, "")
	cfg := &aws.Config{
		Region: aws.String(site.BucketRegion),
		Endpoint:    aws.String(site.Endpoint),
		Credentials: cred,
	}

	sess := session.Must(session.NewSession(cfg))
	svc := s3.New(sess)
	if items, err := getAwsS3ItemMap(svc, site); err != nil {
		logger.Error(err)
	} else {
		logger.Infof("got items: %d, begin delete ...", len(items))
		deleteObjs := &s3.Delete{
			Objects: items,
		}
		d := &s3.DeleteObjectsInput{
			Bucket:                    aws.String(site.Bucket),
			BypassGovernanceRetention: aws.Bool(true),
			Delete:                    deleteObjs,
		}
		if _, err := svc.DeleteObjects(d); err != nil {
			logger.Errorf("delete objects err: %s", err.Error())
		} else {
			logger.Infof("done, deleted items: %d", len(items))
			if _, err := svc.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(site.Bucket)}); err != nil {
				logger.Errorf("delete bucket[%s] failed, err: %s", site.Bucket, err.Error())
			} else {
				logger.Infof("done, deleted bucket: %s", site.Bucket)
			}
		}
	}
	logger.Info("all done")
}

func getAwsS3ItemMap(s3Service *s3.S3, site Site) ([]*s3.ObjectIdentifier, error) {
	var items = make([]*s3.ObjectIdentifier, 0)

	perpage := int64(1000)
	params := &s3.ListObjectsV2Input{
		Bucket:  aws.String(site.Bucket),
		Prefix:  aws.String(site.BucketPath),
		MaxKeys: aws.Int64(perpage),
	}

	logger.Infof("[%s] begin list objects ...", site.Bucket)

	bar := progressbar.Default(100, fmt.Sprintf("list objects [%d/page] ...", perpage))

	npage := 0
	err := s3Service.ListObjectsV2Pages(params,
		func(page *s3.ListObjectsV2Output, last bool) bool {
			logger.Debugf("get page objects: %d", len(page.Contents))
			// Process the objects for each page
			for _, s3obj := range page.Contents {
				items = append(items, &s3.ObjectIdentifier{Key: s3obj.Key})
			}
			npage++
			if npage > bar.GetMax()-1 {
				bar.ChangeMax64(bar.GetMax64() + 100)
			}
			bar.Add(1)
			return true
		},
	)

	bar.Finish()
	logger.Infof("[%s] done list objects", site.Bucket)

	if err != nil {
		// Update errors metric
		logger.Errorf("Error listing %s objects: %s", *params.Bucket, err)
		return nil, err
	}
	return items, nil
}
