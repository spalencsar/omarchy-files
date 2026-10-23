package clipboard

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/duxlabs-dev/dux-go/wlclipboard"
)

type selection interface {
	Set(offers []wlclipboard.Offer) error
	Close()
}

type selector func(choose func([]string) string) ([]byte, string, error)

var imageFormats = []string{"image/png", "image/jpeg", "image/webp"}

var fileFormats = []string{MimeTextUTF8, "text/plain", "UTF8_STRING", "STRING", "TEXT", MimeGnomeCopiedFiles, MimeUriList}

// The owner holds a wayland connection and a goroutine, so it lives only from the first copy until clear or shutdown.
type Service struct {
	newOwner func() (selection, error)
	pick     selector

	mu       sync.Mutex
	owner    selection
	paths    []string
	cut      bool
	plain    string
	unusable bool
}

func NewService() *Service {
	return &Service{
		newOwner: func() (selection, error) { return wlclipboard.NewOwner() },
		pick:     wlclipboard.PasteSelect,
	}
}

func (s *Service) Image() ([]byte, string, bool) {
	data, mime, err := s.pick(chooseImage)
	if err != nil || len(data) == 0 {
		return nil, "", false
	}
	return data, mime, true
}

// An image-only clipboard is declined so no image crosses the pipe just to find no paths in it.
func chooseFiles(offered []string) string {
	for _, want := range fileFormats {
		if slices.Contains(offered, want) {
			return want
		}
	}
	return ""
}

func chooseImage(offered []string) string {
	for _, want := range imageFormats {
		if slices.Contains(offered, want) {
			return want
		}
	}
	for _, mime := range offered {
		if strings.HasPrefix(mime, "image/") {
			return mime
		}
	}
	return ""
}

func Handles(method string) bool { return strings.HasPrefix(method, "clipboard.") }

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drop()
}

func (s *Service) Copy(paths []string, cut bool) error {
	if len(paths) == 0 {
		return errors.New("clipboard.copy requires paths")
	}
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return errors.New("clipboard.copy requires absolute paths: " + path)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner == nil {
		owner, err := s.newOwner()
		if err != nil {
			s.unusable = true
			return err
		}
		s.owner = owner
	}

	plain := string(PlainText(paths))
	if err := s.owner.Set(offers(paths, cut, plain)); err != nil {
		s.drop()
		return err
	}
	s.paths = append([]string(nil), paths...)
	s.cut = cut
	s.plain = plain
	return nil
}

func (s *Service) Paths() ([]string, bool, string) {
	data, mime, err := s.pick(chooseFiles)
	if err != nil {
		return nil, false, ""
	}

	s.mu.Lock()
	mine := s.paths != nil && string(data) == s.plain
	paths, cut := append([]string(nil), s.paths...), s.cut
	s.mu.Unlock()

	if mine {
		return paths, cut, "omarchy-files"
	}
	foreign, foreignCut := Parse(data, mime)
	return foreign, foreignCut, mime
}

func (s *Service) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drop()
}

// Stays true until a copy fails, so the UI never probes wayland at startup.
func (s *Service) Available() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.unusable
}

func (s *Service) drop() {
	if s.owner != nil {
		s.owner.Close()
		s.owner = nil
	}
	s.paths = nil
	s.cut = false
	s.plain = ""
}

func offers(paths []string, cut bool, plain string) []wlclipboard.Offer {
	list := []wlclipboard.Offer{
		{MimeType: MimeGnomeCopiedFiles, Data: GnomePayload(paths, cut)},
		{MimeType: MimeUriList, Data: UriList(paths)},
	}
	if cut {
		list = append(list, wlclipboard.Offer{MimeType: MimeKdeCutSelection, Data: []byte("1")})
	}
	return append(list, wlclipboard.ExpandOffers([]byte(plain), MimeTextUTF8)...)
}
