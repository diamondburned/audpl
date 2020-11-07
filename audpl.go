package audpl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Playlist struct {
	Name   string
	Tracks []Track
}

type Track struct {
	URI,
	Title,
	Artist,
	Album,
	AlbumArtist,
	Comment,
	Genre,
	Year,
	TrackNumber,
	Length,
	Bitrate,
	Codec,
	Quality string
}

func ParseFile(file string) (*Playlist, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open file")
	}
	defer f.Close()

	return Parse(f)
}

func Parse(r io.Reader) (*Playlist, error) {
	scanner := bufio.NewScanner(r)

	if !scanner.Scan() {
		return nil, errors.Wrap(scanner.Err(), "Failed to scan for playlist name")
	}

	k, plname := splitKV(scanner.Text())
	if k != "title" {
		return nil, errors.Errorf("Unexpected playlist header key name %s", k)
	}

	var pl = Playlist{
		Name:   plname,
		Tracks: make([]Track, 0, 2), // arbitrary 2
	}

	var track Track

	for scanner.Scan() {
		switch k, v := splitKV(scanner.Text()); k {
		case "uri":
			// Encountered new URI, signals start of new track. Push the old one
			// into the list, but only if it has a URI. It doesn't have one when
			// we're just starting to parse.
			if track.URI != "" {
				pl.Tracks = append(pl.Tracks, track)
			}

			track.URI = v
		case "title":
			track.Title = v
		case "artist":
			track.Artist = v
		case "album":
			track.Album = v
		case "album-artist":
			track.AlbumArtist = v
		case "genre":
			track.Genre = v
		case "year":
			track.Year = v
		case "track-number":
			track.TrackNumber = v
		case "length":
			track.Length = v
		case "bitrate":
			track.Bitrate = v
		case "codec":
			track.Codec = v
		case "quality":
			track.Quality = v
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "Failed to scan")
	}

	// Push the playlist in one last time.
	pl.Tracks = append(pl.Tracks, track)

	return &pl, nil
}

func splitKV(text string) (string, string) {
	tsplit := strings.SplitN(text, "=", 2)
	if len(tsplit) < 2 {
		return tsplit[0], ""
	}

	q, err := url.QueryUnescape(tsplit[1])
	if err == nil {
		tsplit[1] = q
	}

	return tsplit[0], tsplit[1]
}

func (p Playlist) SaveTo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "title=%s\n", url.QueryEscape(p.Name))
	if err != nil {
		return err
	}

	for _, track := range p.Tracks {
		err := writePair(w,
			"uri", track.URI,
			"title", track.Title,
			"artist", track.Artist,
			"album", track.Album,
			"album-artist", track.AlbumArtist,
			"genre", track.Genre,
			"year", track.Year,
			"track-number", track.TrackNumber,
			"length", track.Length,
			"bitrate", track.Bitrate,
			"codec", track.Codec,
			"quality", track.Quality,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (p Playlist) SaveToBytes() ([]byte, error) {
	var buf bytes.Buffer

	if err := p.SaveTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writePair(w io.Writer, pairs ...string) error {
	for i := 0; i < len(pairs); i += 2 {
		k, v := pairs[i], pairs[i+1]
		if v == "" {
			continue
		}

		_, err := fmt.Fprintf(w, "%s=%s\n", k, url.QueryEscape(v))
		if err != nil {
			return err
		}
	}

	return nil
}
