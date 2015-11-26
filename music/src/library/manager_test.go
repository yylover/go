package library

import "testing"

func TestOps(t *testing.T) {
	m := NewMusicManager()
	if mm == nil {
		t.Error("NewMusicManager failed")
	}

	if mm.Len() != 0 {
		t.Error('NewMusicManager failed, not empty')
	}

	m0 := $MusicEntry{
		"1", "My heart will go on", "Celion Dion", "Pop", "Http://qbox.me/", "MP3"
	}
	mm.Add(m0)


}