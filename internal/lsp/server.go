package lsp

import (
	"bufio"
	"io"

	"github.com/oboard/rune-lang/internal/compiler"
)

func Serve(in io.Reader, out io.Writer) error {
	s := &server{
		reader: bufio.NewReader(in),
		out:    out,
		docs:   map[string]string{},
		cache:  map[string]programCacheEntry{},
	}
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := s.handle(msg); err != nil {
			return err
		}
	}
}

type server struct {
	reader *bufio.Reader
	out    io.Writer
	docs   map[string]string
	cache  map[string]programCacheEntry
}

type programCacheEntry struct {
	text  string
	prog  *compiler.Program
	diags []compiler.Diagnostic
}

func (s *server) setDocument(uri string, text string) {
	if s.docs == nil {
		s.docs = map[string]string{}
	}
	s.docs[uri] = text
	s.clearProgramCache()
}

func (s *server) closeDocument(uri string) {
	delete(s.docs, uri)
	s.clearProgramCache()
}

func (s *server) clearDocuments() {
	s.docs = map[string]string{}
	s.clearProgramCache()
}

func (s *server) clearProgramCache() {
	if s.cache == nil {
		return
	}
	clear(s.cache)
}

func (s *server) analyze(uri string) (*compiler.Program, []compiler.Diagnostic) {
	text := s.docs[uri]
	if s.cache != nil {
		if entry, ok := s.cache[uri]; ok && entry.text == text {
			return entry.prog, entry.diags
		}
	} else {
		s.cache = map[string]programCacheEntry{}
	}
	prog, diags := compiler.AnalyzeSource(uri, text)
	s.cache[uri] = programCacheEntry{text: text, prog: prog, diags: diags}
	return prog, diags
}
