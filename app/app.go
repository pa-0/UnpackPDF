package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func FindFolder(path string, filesMap map[string][]string) {
	folders, err := os.ReadDir(path)
	if err != nil {
		log.Printf("erro ao abrir a pasta %s: %v", path, err.Error())
		return
	}

	for _, folder := range folders {
		if folder.IsDir() {
			folderName := folder.Name()

			if folder.Name() == "guias" || folderName == "laudos" || folderName == "guias - tiss" {
				currentFolder := filepath.Join(path, folder.Name())

				files, err := os.ReadDir(currentFolder)
				if err != nil {
					log.Printf("erro para abrir a pasta %s: %v", currentFolder, err.Error())
					continue
				}

				for _, file := range files {
					filePath := filepath.Join(currentFolder, file.Name())
					if !file.IsDir() && filepath.Ext(file.Name()) == ".pdf" || filepath.Ext(file.Name()) == ".jpg" {

						if filesMap[folderName] == nil {
							filesMap[folderName] = []string{}
						}

						filesMap[folderName] = append(filesMap[folderName], filePath)
					}
				}

			}

			currentFolder := filepath.Join(path, folder.Name())
			FindFolder(currentFolder, filesMap)
		}
	}

}

func MergeJPGsToPDF(filesNameJPG []string, outputFilePath string) error {
	importDefault := pdfcpu.DefaultImportConfig()
	conf := model.NewDefaultConfiguration()

	err := api.ImportImagesFile(filesNameJPG, outputFilePath, importDefault, conf)
	if err != nil {
		return fmt.Errorf("erro ao agrupar os JPGs em PDF: %v", err)
	}
	return nil
}

func TotalSize(files []string) (int64, error) {
	var totalSize int64
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return 0, fmt.Errorf("erro ao obter informações do arquivo %s: %v", file, err)
		}
		totalSize += info.Size()
	}
	return totalSize, nil
}

func MergePDF(path string) {
	filesMap := make(map[string][]string)

	FindFolder(path, filesMap)

	for folder, files := range filesMap {
		if len(files) > 0 {
			totalSize, err := TotalSize(files)
			if err != nil {
				log.Printf("erro para calcular o tamanho total dos arquivos da pasta %s. %v", folder, err.Error())
				continue
			}

			if totalSize > 100*1024*1024 {
				log.Printf("O tamanho total dos arquivos na pasta %s é de %.2f MB. Nenhum arquivo será gerado.\n", folder, float64(totalSize)/1024/1024)
				continue
			}

			fileType := filepath.Ext(files[0])
			newFileName := fmt.Sprintf("%s_merged.pdf", folder)

			if fileType == ".pdf" {
				newFile := filepath.Join(path, newFileName)
				err := api.MergeCreateFile(files, newFile, false, nil)
				if err != nil {
					log.Printf("Erro ao agrupar PDFs da pasta %s: %v", folder, err)
					log.Fatal(1)
					continue
				}
			}

			if fileType == ".jpg" {
				jpgOutputFile := filepath.Join(path, newFileName)
				err := MergeJPGsToPDF(files, jpgOutputFile)
				if err != nil {
					log.Printf("Erro ao agrupar JPGs da pasta %s: %v", folder, err)
					log.Fatal(1)
					continue
				}
			}

			fmt.Printf("O tamanho total dos arquivos na pasta %s é menor que 100 MB\n", folder)
		}
	}

}
