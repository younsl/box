package config

const (
	AssetsPath         = "assets"
	OutputPath         = "generated_punks"
	ImageSizeX         = 24
	ImageSizeY         = 24
	NumPunksToGenerate = 10
)

type PartLayer struct {
	Name     string
	Optional bool
}

var PartLayers = []PartLayer{
	{Name: "background", Optional: true},
	{Name: "punks", Optional: false},
	{Name: "cheek", Optional: true},
	{Name: "beard", Optional: true},
	{Name: "glasses", Optional: true},
	{Name: "top", Optional: true},
	{Name: "mouth", Optional: true},
	{Name: "chain", Optional: true},
}
