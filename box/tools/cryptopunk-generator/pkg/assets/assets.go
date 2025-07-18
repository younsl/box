package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/config"
)

// AllAssets는 로드된 모든 파츠의 파일 경로를 저장합니다.
// 키는 파츠 이름 (예: "punks", "glasses")이고 값은 해당 파츠의 이미지 파일 경로 슬라이스입니다.
var AllAssets = make(map[string][]string)

// LoadAssets는 config.PartLayers에 정의된 모든 파츠에 대한 에셋을 로드합니다.
// raceFilter가 지정되면 해당 종족의 "punks" 에셋만 로드합니다.
func LoadAssets(raceFilter string) error {
	for _, partInfo := range config.PartLayers {
		partPath := filepath.Join(config.AssetsPath, partInfo.Name)
		files, err := getAvailableParts(partPath)
		if err != nil {
			if !partInfo.Optional {
				return fmt.Errorf("failed to load required part '%s': %w", partInfo.Name, err)
			}
			// Optional 파츠는 에러를 무시하고 계속 진행합니다.
			fmt.Printf("Warning: No assets found for optional part '%s', skipping. Error: %v\n", partInfo.Name, err)
			AllAssets[partInfo.Name] = []string{}
			continue
		}

		if partInfo.Name == "punks" && raceFilter != "" {
			filteredFiles := filterPunkAssetsByRace(files, raceFilter, partPath)
			AllAssets[partInfo.Name] = filteredFiles
			if len(filteredFiles) == 0 {
				fmt.Printf("Warning: No punks found for race '%s' in %s. No punks of this type will be generated.\n", raceFilter, partPath)
			} else {
				fmt.Printf("Filtered punks by race: %s. Found %d assets in %s.\n", raceFilter, len(filteredFiles), partPath)
			}
		} else {
			AllAssets[partInfo.Name] = files
		}
	}
	return nil
}

// filterPunkAssetsByRace는 주어진 파일 목록에서 특정 raceFilter를 포함하는 "punks" 에셋만 필터링합니다.
func filterPunkAssetsByRace(files []string, raceFilter string, partPath string) []string {
	var filteredFiles []string
	for _, file := range files {
		// 파일명의 대소문자를 구분하지 않고 raceFilter 문자열이 포함되어 있는지 확인합니다.
		if strings.Contains(strings.ToLower(filepath.Base(file)), strings.ToLower(raceFilter)) {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles
}

// getAvailableParts는 지정된 경로에서 사용 가능한 이미지 파일(.png) 목록을 반환합니다.
func getAvailableParts(partsPath string) ([]string, error) {
	if _, err := os.Stat(partsPath); os.IsNotExist(err) {
		// 디렉토리가 없으면 빈 슬라이스를 반환합니다 (옵션 파츠일 수 있으므로 에러는 아님).
		return []string{}, nil
	}

	var imageFiles []string
	dirEntries, err := os.ReadDir(partsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", partsPath, err)
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".png" {
			imageFiles = append(imageFiles, filepath.Join(partsPath, entry.Name()))
		}
	}
	return imageFiles, nil
}
