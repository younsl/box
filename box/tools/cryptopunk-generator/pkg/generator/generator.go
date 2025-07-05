package generator

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	_ "image/png" // PNG 포맷 디코더 등록
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/assets"
	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/config"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // 랜덤 시드 초기화
}

// GeneratePunk는 랜덤 파츠를 조합하여 펑크 이미지를 생성하고 저장합니다.
func GeneratePunk(punkID int, raceFilter string) error {
	baseImage := image.NewRGBA(image.Rect(0, 0, config.ImageSizeX, config.ImageSizeY))
	punkRaceName := "unknown"
	isFemalePunk := false

	// 디렉토리 준비
	if _, err := os.Stat(config.OutputPath); os.IsNotExist(err) {
		err := os.MkdirAll(config.OutputPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	for _, partInfo := range config.PartLayers {
		partName := partInfo.Name
		isOptional := partInfo.Optional

		partsList, ok := assets.AllAssets[partName]
		if !ok || len(partsList) == 0 {
			if !isOptional {
				fmt.Printf("Warning: Required part '%s' has no assets.\n", partName)
			}
			continue
		}

		if isOptional && rand.Float32() < 0.5 { // 50% 확률로 옵션 파츠 건너뛰기
			// fmt.Printf("Skipping optional part: %s\n", partName) // 필요시 주석 해제
			continue
		}

		if isFemalePunk && partName == "chain" {
			fmt.Printf("Skipping chain for female punk: %s\n", punkRaceName)
			continue
		}

		partImagePath := partsList[rand.Intn(len(partsList))] // 랜덤 선택
		partImageFile, err := os.Open(partImagePath)
		if err != nil {
			fmt.Printf("Warning: Failed to open part image '%s': %v. Skipping.\n", partImagePath, err)
			continue
		}
		defer partImageFile.Close()

		img, _, err := image.Decode(partImageFile)
		if err != nil {
			fmt.Printf("Warning: Failed to decode part image '%s': %v. Skipping.\n", partImagePath, err)
			continue
		}

		if partName == "punks" {
			punkRaceName = strings.TrimSuffix(filepath.Base(partImagePath), filepath.Ext(partImagePath))
			if strings.Contains(strings.ToLower(punkRaceName), "female") {
				isFemalePunk = true
			}
		}

		// 이미지 크기가 다를 경우 리사이즈 (24x24로 강제, Nearest Neighbor 방식은 draw.Draw로 구현)
		if img.Bounds().Dx() != config.ImageSizeX || img.Bounds().Dy() != config.ImageSizeY {
			// 단순 복사로 리사이징 효과 (Pillow의 NEAREST와 정확히 동일하지 않을 수 있음)
			// Go에서는 고급 리사이징 라이브러리 (예: github.com/nfnt/resize) 사용 고려 가능
			tempImage := image.NewRGBA(image.Rect(0, 0, config.ImageSizeX, config.ImageSizeY))
			draw.Draw(tempImage, tempImage.Bounds(), img, image.Point{}, draw.Src) // Src는 단순 복사
			img = tempImage
			fmt.Printf("Resized part '%s' from %dx%d to %dx%d\n", partName, img.Bounds().Dx(), img.Bounds().Dy(), config.ImageSizeX, config.ImageSizeY)
		}

		draw.Draw(baseImage, baseImage.Bounds(), img, image.Point{}, draw.Over)
	}

	outputFilename := filepath.Join(config.OutputPath, fmt.Sprintf("punk_%s_%03d.png", punkRaceName, punkID))
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", outputFilename, err)
	}
	defer outputFile.Close()

	encoder := png.Encoder{CompressionLevel: png.NoCompression} // 필요시 압축 레벨 조절
	if err := encoder.Encode(outputFile, baseImage); err != nil {
		return fmt.Errorf("failed to encode image to '%s': %w", outputFilename, err)
	}

	fmt.Printf("Generated: %s\n", outputFilename)
	return nil
}
