package app

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func FindFolder(dir string) {

	folders, err := os.ReadDir(dir)
	if err != nil {
		log.Default().Printf("erro para abrir a pasta tmp. %v", err.Error())
		return
	}

	for _, folder := range folders {

		if folder.IsDir() {
			if folder.Name() == "guias" || folder.Name() == "laudos" || folder.Name() == "guias - tiss" {

				currentFolder := filepath.Join(dir, folder.Name())
				//fmt.Println(currentFolder)

				// Remove arquivos antigos antes de criar novos
				err := removeOldMergedFiles(currentFolder)
				if err != nil {
					log.Default().Printf("Erro ao tentar remover arquivos antigos na pasta %s: %v", currentFolder, err)
					continue
				}

				files, err := os.ReadDir(currentFolder)
				if err != nil {
					log.Default().Printf("erro para abrir a pasta %s. %v", currentFolder, err.Error())
					continue
				}

				var fileSize int64 = 0
				var fileNames []string
				var typeFile string

				for _, file := range files {

					if !file.IsDir() && filepath.Ext(file.Name()) == ".pdf" || filepath.Ext(file.Name()) == ".jpg" {

						fileName := filepath.Join(currentFolder, file.Name())
						//fmt.Println(fileName)
						fileNames = append(fileNames, fileName)

						fileInfo, err := file.Info()
						if err != nil {
							log.Default().Printf("erro ao obter as informações do arquivo %s. %v", file.Name(), err.Error())
							continue
						}

						fileSize += fileInfo.Size()
						typeFile = filepath.Ext(file.Name())
					}
				}

				if len(fileNames) == 0 {
					continue
				}

				totalSizeMB := float64(fileSize) / (1024 * 1024)
				if totalSizeMB > 100000 {
					fmt.Printf("Tamanho total dos arquivos %s é %.2fMB", fileNames, totalSizeMB)
					time.Sleep(3 * time.Second)
					continue
				}

				switch typeFile {
				case ".pdf":

					err = mergeFiles(fileNames, currentFolder)
					if err != nil {
						log.Default().Printf("erro ao tentar mesclar os arquivos")
						time.Sleep(3 * time.Second)
						continue
					}
				case ".jpg":

					outPutFile := filepath.Join(currentFolder, "laudos_merged.jpg")
					err := mergeJPGsToImage(fileNames, outPutFile)
					if err != nil {
						log.Default().Printf("Erro ao fazer o merge dos JPGs: %v", err)
						time.Sleep(3 * time.Second)
						continue
					}

					err = mergeJPGToPDF(fileNames, currentFolder)
					if err != nil {
						log.Default().Printf("Erro ao fazer o merge dos JPGs para PDF: %v", err)
						time.Sleep(3 * time.Second)
						continue
					}

				}

			} else {

				currentFolder := filepath.Join(dir, folder.Name())
				//recursive
				FindFolder(currentFolder)
			}

		}
	}
}

// mergeFiles somente para arquivos PDF - Usando a api PDFCPU
func mergeFiles(fileNames []string, currentFolder string) error {
	tempFile := filepath.Join(currentFolder, "merged_temp.pdf")
	err := api.MergeCreateFile(fileNames, tempFile, false, nil)
	if err != nil {
		log.Default().Printf("erro ao tentar criar o arquivo temporário %s. %v", tempFile, err.Error())
		return err
	}
	fmt.Printf("Arquivo temporário %s criado!\n", tempFile)

	finalFile := filepath.Join(currentFolder, "merged.pdf")
	err = os.Rename(tempFile, finalFile)
	if err != nil {
		log.Default().Printf("erro ao renomear o arquivo temporário %s para %s. %v", tempFile, finalFile, err.Error())
		return err
	}
	fmt.Printf("Arquivo final %s criado com sucesso!\n", finalFile)

	return nil
}

// mergeJPGToPDF - Cria um PDF a partir de arquivos JPG usando PDFCPU
func mergeJPGToPDF(fileNames []string, currentFolder string) error {
	tempFile := filepath.Join(currentFolder, "merged_temp.pdf")

	importDefault := pdfcpu.DefaultImportConfig()
	conf := model.NewDefaultConfiguration()

	fmt.Printf("Tentando criar o arquivo temporário %s...\n", tempFile)
	err := api.ImportImagesFile(fileNames, tempFile, importDefault, conf)
	if err != nil {
		log.Default().Printf("Erro ao tentar criar o arquivo temporário %s: %v", tempFile, err.Error())
		return err
	}
	fmt.Printf("Arquivo temporário %s criado com sucesso!\n", tempFile)

	finalFile := filepath.Join(currentFolder, "merged.pdf")
	fmt.Printf("Tentando renomear o arquivo temporário %s para %s...\n", tempFile, finalFile)

	err = os.Rename(tempFile, finalFile)
	if err != nil {
		log.Default().Printf("Erro ao renomear o arquivo temporário %s para %s: %v", tempFile, finalFile, err.Error())
		return err
	}
	fmt.Printf("Arquivo final %s criado com sucesso!\n", finalFile)

	return nil
}

// mergeJPFsToImagem usado pra mescar os arquivos em jpg - fonte: https://pkg.go.dev/image
func mergeJPGsToImage(jpgFiles []string, output string) error {
	var images []image.Image
	var totalWidth, totalHeight int

	// Carregar todas as imagens e calcular a largura e altura total da imagem combinada
	for _, jpgFile := range jpgFiles {
		file, err := os.Open(jpgFile)
		if err != nil {
			return fmt.Errorf("erro ao abrir o arquivo JPG %s: %w", jpgFile, err)
		}
		defer file.Close()

		img, err := jpeg.Decode(file)
		if err != nil {
			return fmt.Errorf("erro ao decodificar o arquivo JPG %s: %w", jpgFile, err)
		}

		images = append(images, img)
		totalWidth = max(totalWidth, img.Bounds().Dx())
		totalHeight += img.Bounds().Dy()
	}

	// Criar uma nova imagem com as dimensões calculadas
	combinedImage := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	// Desenhar as imagens uma abaixo da outra na imagem combinada
	yOffset := 0
	for _, img := range images {
		draw.Draw(combinedImage, image.Rect(0, yOffset, img.Bounds().Dx(), yOffset+img.Bounds().Dy()), img, image.Point{}, draw.Src)
		yOffset += img.Bounds().Dy()
	}

	// Criar o arquivo de saída
	outputFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("erro ao criar o arquivo de saída %s: %w", output, err)
	}
	defer outputFile.Close()

	// Salvar a imagem combinada no arquivo de saída
	err = jpeg.Encode(outputFile, combinedImage, nil)
	if err != nil {
		return fmt.Errorf("erro ao codificar a imagem combinada: %w", err)
	}

	return nil
}

// Função auxiliar para obter o valor máximo entre dois inteiros
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// DELETAR OS ARQUIVOS MESCLADOS ANTES DE CRIAR OS NOVOS
func removeOldMergedFiles(folder string) error {
	filesToRemove := []string{
		filepath.Join(folder, "merged.pdf"),
		filepath.Join(folder, "merged.jpg"),
	}

	for _, file := range filesToRemove {
		if fileExists(file) {
			err := os.Remove(file)
			if err != nil {
				log.Default().Printf("Erro ao remover o arquivo %s: %v", file, err)
				return err
			}
			fmt.Printf("Arquivo %s removido com sucesso.\n", file)
		}
	}
	return nil
}

// Função auxiliar para verificar se um arquivo existe
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
