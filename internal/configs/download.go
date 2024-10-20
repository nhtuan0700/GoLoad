package configs

type DownloadMode string

const (
	DownloadModeLocal DownloadMode = "local"
	DownloadModelS3   DownloadMode = "s3"
)

type Download struct {
	Mode              DownloadMode `yaml:"mode"`
	Bucket            string       `yaml:"bucket"`
	Address           string       `yaml:"address"`
	Username          string       `yaml:"username"`
	Password          string       `yaml:"password"`
	DownloadDirectory string       `yaml:"download_directory"`
}
