package camera

import "testing"

//import "fmt"

func TestVideoFiles(t *testing.T) {
	store := NewVideoStorage("/home/csoria/tmp/cam")
	videos := store.VideoFiles()
	if len(videos) == 0 {
		t.Errorf("Expecting non-zero video files, got (%v)\n", len(videos))
	}
}
