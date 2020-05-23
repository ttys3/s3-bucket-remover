# s3-bucket-remover

## usage

```bash
s3-bucket-remover -k ACCESS_KEY -s ACCESS_SECRET -e ENDPOINT -b BUCKET_NAME
```

## available params

```bash
❯ ./s3-bucket-remover -h
Usage of ./s3-bucket-remover:
  -b string
    	bucket to remove
  -e string
    	endpoint
  -k string
    	access Key
  -l string
    	log level (default "info")
  -p string
    	bucket path prefix to remove (default "/")
  -r string
    	region (default "us-east-1")
  -s string
    	secret
  -v	show app version
```