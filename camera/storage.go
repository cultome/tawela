package camera

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

type Video struct {
	Path     string
	Filename string
	Date     time.Time
}

type VideoStorage struct {
	Path  string
	regex *regexp.Regexp
}

func NewVideoStorage(storagePath string) *VideoStorage {
	dir, _ := os.Open(storagePath)
	defer dir.Close()

	stat, _ := dir.Stat()

	if !stat.IsDir() {
		panic(fmt.Sprintf("File [%v] is not a directory.", storagePath))
	}

	regex, _ := regexp.Compile(VideoFilenameRegexp)

	return &VideoStorage{storagePath, regex}
}

func (video *Video) String() string {
	return fmt.Sprintf("%v [%v]", video.Filename, video.Date)
}

func (storage *VideoStorage) VideoFiles() []Video {
	dir, _ := os.Open(storage.Path)
	defer dir.Close()

	files, _ := dir.Readdirnames(100)
	videos := []Video{}

	for _, file := range files {
		groups := storage.regex.FindStringSubmatch(file)
		if groups != nil {
			video := storage.NewVideo(groups[1:], file, storage.Path)
			videos = append(videos, video)
		}
	}
	return videos
}

func (storage *VideoStorage) NewVideo(dateStr []string, filename, filepath string) Video {
	date := make([]int, 6)
	for idx, g := range dateStr {
		date[idx], _ = strconv.Atoi(g)
	}

	loc, _ := time.LoadLocation("America/Mexico_City")
	month := storage.parseMonth(date[1])

	video := Video{
		path.Join(filepath, filename),
		filename,
		time.Date(date[0]+2000, month, date[2], date[3], date[4], date[5], 0, loc),
	}
	return video
}

func (storage *VideoStorage) parseMonth(monthNbr int) time.Month {
	switch monthNbr {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	case 12:
		return time.December
	}
	return 0
}
