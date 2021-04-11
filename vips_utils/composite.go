package vips_utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

func CompositeImgFromFiles(files []string, outFile string) error {

	if len(files) <= 1 {
		return errors.New("composite files length is <= 1")
	}

	// boot file
	bootFile := make([]string, 0, len(files))

	// gif file
	var gifFile string

	// find gif
	for _, f := range files {

		if strings.HasSuffix(f, ".gif") {
			if gifFile != "" {
				return errors.New("not support multi gif")
			}
			gifFile = f
			continue
		}

		bootFile = append(bootFile, f)
	}

	log.Println("bootFile", bootFile)
	boot, err := compositeBootFiles(bootFile)
	if err != nil {
		return err
	}

	// has gif
	if len(gifFile) > 0 {
		log.Println("gifFile", gifFile)
		// open gif
		gif, err := vips.NewImageFromFile(gifFile)
		if err != nil {
			return err
		}

		// composite boot and gif
		result, err := compositeBootAndGif(boot, gif)
		if err != nil {
			return err
		}

		buf, _, err := result.ExportGif(nil)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(outFile, buf, 0644)
	}

	buf, _, err := boot.ExportNative()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outFile, buf, 0644)
}

func compositeBootFiles(files []string) (*vips.ImageRef, error) {

	if len(files) == 0 {
		return nil, errors.New("composite boot files length is <= 1")
	}

	boot, err := vips.NewImageFromFile(files[0])
	if err != nil {
		return nil, err
	}

	images := make([]*vips.ImageComposite, 0, len(files))
	for i := 1; i < len(files); i++ {
		image, err := vips.NewImageFromFile(files[i])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("open file %s error err:%s", files[i], err.Error()))
		}
		images = append(images, &vips.ImageComposite{image, vips.BlendModeOver, 0, 0})
	}

	if len(images) > 0 {
		err = boot.CompositeMulti(images)
		if err != nil {
			return nil, err
		}
	}

	return boot, nil
}

func compositeBootAndGif(boot *vips.ImageRef, gif *vips.ImageRef) (*vips.ImageRef, error) {

	// page height
	pageHeight, err := gif.GetInt("page-height")
	if err != nil {
		return nil, err
	}

	// n-page
	nPages, err := gif.GetIntDefault("n-pages", 0)
	if err != nil {
		return nil, err
	}

	// frames
	frames := make([]*vips.ImageRef, nPages)
	for pageNum := 0; pageNum < nPages; pageNum++ {

		newGif, err := gif.Copy()
		if err != nil {
			return nil, err
		}

		err = newGif.ExtractArea(0, pageNum*pageHeight, gif.Width(), pageHeight)
		if err != nil {
			return nil, err
		}

		newFrame, err := boot.Copy()
		if err != nil {
			return nil, err
		}

		err = newGif.Composite(newFrame, vips.BlendModeDestOver, 0, 0)
		if err != nil {
			return nil, err
		}

		frames[pageNum] = newGif

	}

	err = boot.Arrayjoin(frames)
	if err != nil {
		return nil, err
	}

	return boot, nil
}
