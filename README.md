# selfupdater
[![Go Reference](https://pkg.go.dev/badge/github.com/tomasen/selfupdater.svg)](https://pkg.go.dev/github.com/tomasen/selfupdater)

Build self-updating golang programs with customizable over-the-air(OTA) updates protocol.

`SelfUpdater` performs following tasks in sequence:
1. Calculating local checksum of the current executable file.
2. Fetch the remote checksum.
3. Compare the local checksum with the remote checksum, proceed if there is a difference.
4. Download remote executable file to a temporary file.
5. Calculating the checksum of the downloaded temporary file, proceed if it's the same to 
   the remote checksum we obtained on step 2.
6. Overwrite the current executable file with the downloaded temporary file. 
7. Cleanup the temporary file.
8. Safely restart the current executable program by using [rollover](https://pkg.go.dev/github.com/tomasen/rollover).

## Customize Update Source

Implement `UpdateProvider` interface to establish customized updates protocol.

`S3UpdateProvider` is an out-of-the-box example of a customized `UpdateProvider` based 
on AWS S3 and md5 checksum.
Call `NewS3UpdateProvider()` to specify the S3 bucket and region, 
paths of the executable and checksum file and create the UpdateProvider, 
feed it to `NewSelfUpdate(provider UpdateProvider)`. 

## Unit Test

1. Build the base version of the executable file.

   `go build -o ./bin/selfupdater ./example/base`

2. Build the update version of the executable file.
   
   `go build -o ./bin/selfupdater-update ./example/update`
   
3. Calculate the checksum and push the updates to S3 by using `awscli`.

   ```
   md5 -q ./bin/selfupdater-update | tr -d '\n' > ./bin/selfupdater-update.md5
   aws s3 cp ./bin/selfupdater-update  s3://selfupdater-tomasen/selfupdater
   aws s3 cp ./bin/selfupdater-update.md5  s3://selfupdater-tomasen/selfupdater.md5
   ```

4. Run the base version.
   
   `./bin/selfupdater`

5. The base is expected to be updated to the update version, with stdout like
   ```
   ...
   this is the base version
   this is the base version
   2021/03/12 03:39:11 child started running, pid: 53466
   this is the update version
   this is the update version
   ...
   ```
   
6. Run the update version

   `./bin/selfupdater-update`
   
   should expect stdout something similar likes:

   ```
   this is the update version
   this is the update version
   update finished
   this is the update version
   this is the update version
   ...
   ```   