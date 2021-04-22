package image

func DownloadImage(imageUrl, savePath, cachePath string) error {
	operator, err := NewImageOperator(imageUrl, savePath, cachePath)
	if err != nil {
		return err
	}
	if err := operator.SaveImage(); err != nil {
		return err
	}
	return nil
}

func DownloadFile(imageUrl, savePath, cachePath string) error {
	operator, err := NewImageOperator(imageUrl, savePath, cachePath)
	if err != nil {
		return err
	}
	return operator.SaveLayersFiles()
}
