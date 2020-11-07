package audpl

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	pl := testdata(t, "1000.audpl")

	p, err := Parse(pl)
	if err != nil {
		t.Fatal(err)
	}

	testPlaylist(t, p)
}

func testPlaylist(t *testing.T, p *Playlist) {
	if p.Name != "HYDE" {
		t.Fatalf("Unexpected playlist name, got %q != expected %q", p.Name, "HYDE")
	}

	var tracks = []string{
		"WHO'S GONNA SAVE US",
		"MAD QUALIA",
		//    v intentional
		"SICK（feat. Matt B of From Ashes to New)",
		"ANOTHER MOMENT",
		"FAKE DIVINE",
		"AFTER LIGHT",
		"OUT",
		"ZIPANG (feat. YOSHIKI)",
		"SET IN STONE",
		"LION",
		"TWO FACE",
		"MIDNIGHT CELEBRATION II anti mix",
		"ORDINARY WORLD",
	}

	if got, expected := len(p.Tracks), len(tracks); got != expected {
		t.Fatalf("Mismatch length, got %d != expected %d", got, expected)
	}

	for i, track := range p.Tracks {
		if track.Title != tracks[i] {
			t.Errorf("Unexpected track %d, got %q != expected %q", i, track.Title, tracks[i])
		}
	}
}

func TestSave(t *testing.T) {
	pl := testdata(t, "1000.audpl")

	p, err := Parse(pl)
	if err != nil {
		t.Fatal(err)
	}

	rendered, err := p.SaveToBytes()
	if err != nil {
		t.Fatal(err)
	}

	r, err := Parse(bytes.NewReader(rendered))
	if err != nil {
		t.Fatal(err)
	}

	testPlaylist(t, r)
}

func TestSplitKV(t *testing.T) {
	var tests = []struct{ input, k, v string }{
		{
			`title=WHO%27S%20GONNA%20SAVE%20US`,
			"title", "WHO'S GONNA SAVE US",
		},
		{
			`uri=/HYDE%20-%20Anti%20%5B320K%5D/02.%20MAD%20QUALIA.mp3`,
			"uri", "/HYDE - Anti [320K]/02. MAD QUALIA.mp3",
		},
		{
			`quality=Stereo%2C%2044100%20Hz`,
			"quality", "Stereo, 44100 Hz",
		},
		{
			`k=03%20%E7%B5%82%E3%82%8F%E3%82%8A%E3%81%AE%E3%81%AF%E3%81%97%E3%82%99%E3%81%BE%E3%82%8A.flac`,
			"k", "03 終わりのはじまり.flac",
		},
	}

	for _, test := range tests {
		k, v := splitKV(test.input)

		if k != test.k {
			t.Fatalf("Unexpected key %q, expected %q", k, test.k)
		}
		if v != test.v {
			t.Fatalf("Unexpected value %q, expected %q", v, test.v)
		}
	}

	// Test invalid.
	k, v := splitKV("invalid input")
	if v != "" {
		t.Fatal("Unexpected value for invalid input:", k, v)
	}
}

func testdata(t *testing.T, file string) io.Reader {
	t.Helper()

	f, err := os.Open(filepath.Join("_testdata", file))
	if err != nil {
		t.Fatal("Failed to open test data:", err)
	}

	t.Cleanup(func() { f.Close() })
	return f
}
