package service

import (
	"os"
)

func InitCachePath(id string) error {
	WorkPath, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.MkdirAll(WorkPath+"/cache/"+id, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func MergeFile(tsPath, fileName string) error {
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer outFile.Close()

	tsFileList, err := os.ReadDir(tsPath)
	if err != nil {
		return err
	}

	for _, f := range tsFileList {
		tsFilePath := tsPath + "/" + f.Name()
		tsFileContent, err := os.ReadFile(tsFilePath)
		if err != nil {
			return err
		}

		if _, err := outFile.Write(tsFileContent); err != nil {
			return err
		}

		if err = os.Remove(tsFilePath); err != nil {
			return err
		}
	}

	err = os.RemoveAll(tsPath)
	if err != nil {
		return err
	}

	WorkPath, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.RemoveAll(WorkPath + "/cache")
	if err != nil {
		return err
	}

	return nil
}
