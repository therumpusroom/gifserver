package gifserver

import (
	"fmt"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

type converter func(string, int) (string, error)

func checkDimensions(reader io.Reader, maxWidth, maxHeight int) error {
	data, err := gif.DecodeConfig(reader)

	if err != nil {
		return err
	}

	if maxWidth > 0 && data.Width > maxWidth {
		return fmt.Errorf("Image width too large %d > %d", data.Width, maxWidth)
	}

	if maxHeight > 0 && data.Height > maxHeight {
		return fmt.Errorf("Image height too large %d > %d", data.Height, maxHeight)
	}

	return nil
}

// convert -coalesce brocoli.gif out%05d.pgm

func extractGif(dir string) error {
	log.Print("Extracting ", dir)
	pattern := "frame_%05d.png"

	cmd := exec.Command("convert",
		"-coalesce",
		"in.gif",
		pattern)

	cmd.Dir = dir
	return cmd.Run()
}

// ffmpeg -i "$pattern" -pix_fmt yuv420p -vf 'scale=trunc(in_w/2)*2:trunc(in_h/2)*2' "${out_base}.mp4"

func convertToMP4(dir string, frameRate int) (string, error) {
	log.Print("Encoding ", dir, " to mp4")

	outFname := "out.mp4"

	options := getOptions(frameRate)
	options = append(options, "-vf", "scale=trunc(in_w/2)*2:trunc(in_h/2)*2'", outFname)

	cmd := exec.Command("ffmpeg", options...)

	cmd.Dir = dir
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return path.Join(dir, outFname), nil
}

// ffmpeg -i "$pattern" -q 5 -pix_fmt yuv420p "${out_base}.ogv"

func convertToOGV(dir string, frameRate int) (string, error) {
	log.Print("Encoding ", dir, " to ogv")

	outFname := "out.ogv"

	options := getOptions(frameRate)
	options = append(options, "-q", "5", outFname)

	cmd := exec.Command("ffmpeg", options...)

	cmd.Dir = dir
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return path.Join(dir, outFname), nil
}

func convertToFrame(dir string, frameRate int) (string, error) {
	return path.Join(dir, "frame_00001.png"), nil
}

func getOptions(frameRate int) []string {
	options := []string{
		"-i", "frame_%05d.png",
		"-pix_fmt", "yuv420p",
	}

	if frameRate != 0 {
		options = append(options, "-r", fmt.Sprintf("%v", frameRate))
	}

	return options
}

func cleanDir(dir string) error {
	log.Print("Removing ", dir)
	return os.Remove(dir)
}

func prepareConversion(reader io.Reader) (string, error) {
	dir, err := ioutil.TempDir("", "gifserver")

	if err != nil {
		return "", err
	}

	output, err := os.Create(path.Join(dir, "in.gif"))

	if err != nil {
		cleanDir(dir)
		return "", err
	}

	defer output.Close()

	_, err = io.Copy(output, reader)

	if err != nil {
		cleanDir(dir)
		return "", err
	}

	return dir, nil
}

func copyFile(src, dest string) error {
	log.Print("Copying ", src, " to ", dest)

	input, err := os.Open(src)
	if err != nil {
		return err
	}

	output, err := os.Create(dest)

	if err != nil {
		return err
	}

	defer output.Close()

	_, err = io.Copy(output, input)

	if err != nil {
		return err
	}

	defer input.Close()
	return nil
}
