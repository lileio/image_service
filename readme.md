# ImageService
[![Build Status](https://travis-ci.org/lileio/image_service.svg?branch=master)](https://travis-ci.org/lileio/image_service)

An gRPC based parallel image processing and storage microservice. Image Service takes an original image and a set of 'operations' to process and streams the results as they're complete back to the client. Optional no operations can be requested and the original is stored using the storage provided for later use with external services (imgix.com etc).

Processing is done by a set of 'workers' to limit concurrency and overloading. 

By default 5 workers are run, meaning only 5 image resizing or uploading operations happen concurrently. New operations are queued and streamed back to the client when completed.

Image processing uses [bimg](https://github.com/h2non/bimg) which is backed by [libvips](http://www.vips.ecs.soton.ac.uk/index.php?title=Libvips).

``` protobuf
service ImageService {
  rpc Store(ImageStoreRequest) returns (stream Image) {}
  rpc Delete(DeleteRequest) returns (DeleteResponse) {}
}
```

## Docker

A pre build Docker container is available at:

```
docker pull lileio/image_service
```

## Setup

Configuration of storage is done via ENV variables

### Worker Pool Size

To increase or decrease the number of parallel workers, set the `WORKER_POOL_SIZE` var.

```
WORKER_POOL_SIZE=10
```

### Cloud Storage

To point to a [cloud storage service](https://github.com/lileio/cloud_storage_service) instance.

```
CLOUD_STORAGE_ADDR=10.0.0.1:8000
```

### File Storage

File system storage is also available, note that you will be responsible for hosting/serving the images and full URL generation.

```
FILE_LOCATION=/path/to/webserver
```

## Example

``` go

// Setup a client connection
client = image_service.NewImageServiceClient(conn)

// Create a request object
b, err := ioutil.ReadFile("pic.jpg")
if err != nil {
  panic(err) 
}

ctx := context.Background()
req := &image_service.ImageStoreRequest{
  Filename: "pic.jpg",
  Data:     b,
  Ops: []*image_service.ImageOperation{
    &image_service.ImageOperation{
      Crop:        true,
      Width:       200,
      Height:      200,
      VersionName: "small",
    },
    &image_service.ImageOperation{
      Crop:        true,
      Width:       600,
      Height:      600,
      VersionName: "medium",
    },
    &image_service.ImageOperation{
      Crop:        true,
      Width:       1000,
      Height:      1000,
      VersionName: "large",
    },
  },
}


stream, err := client.Store(ctx, req)

for {
  img, err := stream.Recv()
  if err == io.EOF {
    break
  }

  if err != nil {
    fmt.Printf("err = %+v\n", err)
  }

  fmt.Printf("img = %+v\n", img.URL)
}
```
