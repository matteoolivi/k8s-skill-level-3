package imageblur

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/matteoolivi/img-blurring-exercise/pkg/helper/rabbitmq"
	"github.com/matteoolivi/img-blurring-exercise/pkg/helper/s3"
	"github.com/streadway/amqp"
)

func Run(stopCh <-chan struct{}, tmpImageFile string) {
	urls, cleanUpRabbitMQConnection, err := rabbitmq.GetURLsChannel()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanUpRabbitMQConnection()

	go func() {
		for url := range urls {
			if err := processURL(url, tmpImageFile); err != nil {
				log.Fatal(err)
			}
		}
	}()

	<-stopCh
	log.Print("Stop signal received: terminating.")
}

func processURL(url amqp.Delivery, imageFilePath string) error {
	imageURL := string(url.Body)
	log.Printf("Processing image URL %s", imageURL)

	// TODO: This is fragile, it means the image name cannot contain "/". Also, it's ugly.
	// There are probably other libraries or abstractions in the same library we use that allow us
	// to download the image only with the URL (i.e. no need to extract the name). Use them instead.
	urlTokens := strings.Split(imageURL, "/")
	imageName := urlTokens[len(urlTokens)-1]

	if err := saveImageToFile(imageName, imageFilePath); err != nil {
		return err
	}

	if err := blurImage(imageFilePath); err != nil {
		return err
	}

	if err := uploadBlurredImage(imageName); err != nil {
		return err
	}

	// Tell RabbitMQ we're done processing the image.
	url.Ack(false)

	log.Printf("Done processing image URL %s", imageURL)
	return nil
}

func saveImageToFile(imageURL, imageFilePath string) error {
	downloader, err := s3.GetDownloader()
	if err != nil {
		return err
	}

	imageFile, err := os.Create(imageFilePath)
	if err != nil {
		return err
	}
	defer imageFile.Close()
	log.Printf("Opened file %s to save image at URL %s", imageFilePath, imageURL)

	err = s3.DownloadImage(imageURL, downloader, *imageFile)
	if err != nil {
		return err
	}

	return nil
}

// TODO: this hardcoded here is horrible, change it to something decent.
const imageBlurCmdTemplate = "/scripts/imageblur/yolo_opencv.py --image ${image_file_to_blur} " +
	"--config /scripts/imageblur/yolov3.cfg --weights /scripts/imageblur/yolov3.weights --classes /scripts/imageblur/yolov3.txt"

func blurImage(imageFilePath string) error {
	expandedCmdStr := strings.Replace(imageBlurCmdTemplate, "${image_file_to_blur}", imageFilePath, 1)
	// TODO: "python3" hardcoded here is horrible, change it to something decent.
	cmd := exec.Command("python3", strings.Split(expandedCmdStr, " ")...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to blur image %s: %s: %s", imageFilePath, err, output)
	}
	log.Printf("Successfully blurred image in %s.", imageFilePath)

	return nil
}

// TODO: This hardcoding here is horrible. Also, the fact that the name of the blurred image is
// constant is even more horrible. Notice that fixing this requires changing the python script that
// is given.
const blurredImageFilePath = "/tmp/filtered.jpg"

func uploadBlurredImage(imageName string) error {
	blurredImageFile, err := os.Open(blurredImageFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := blurredImageFile.Close(); err != nil {
			log.Printf("Failed to close blurred image file %s for image %s: %s",
				blurredImageFilePath,
				imageName,
				err)
		}
	}()

	_, err = s3.UploadImage(imageName, blurredImageFile)
	if err != nil {
		return err
	}

	return nil
}
