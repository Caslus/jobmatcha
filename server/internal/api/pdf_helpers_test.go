package api

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractPDFURLs(t *testing.T) {
	data := []byte(`
/URI(https://example.test/jobs/1)
/URI(mailto:ada@example.test)
/URI/URI(https://example.test/jobs/2)
/URI(https://example.test/jobs/1)
/URI(ftp://example.test/ignored)
/URI(https://example.test/with\(parentheses\))
`)

	got := extractPDFURLs(data)
	want := []string{
		"https://example.test/jobs/1",
		"mailto:ada@example.test",
		"https://example.test/jobs/2",
		`https://example.test/with\(parentheses\)`,
	}
	if len(got) != len(want) {
		t.Fatalf("URL count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("URL %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDecodePDFHexString(t *testing.T) {
	for _, tt := range []struct {
		name string
		hex  string
		want string
	}{
		{name: "UTF-16BE with BOM and whitespace", hex: "FEFF 0041 0064 0061 000A", want: "\ufeffAda"},
		{name: "plain bytes", hex: "416461", want: "Ada"},
		{name: "odd length is padded", hex: "f", want: "\x0f"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodePDFHexString(tt.hex); got != tt.want {
				t.Errorf("decodePDFHexString(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}

func TestExtractResumeText(t *testing.T) {
	pdf := testPDF("Ada Lovelace", "https://example.test/portfolio")

	for _, tt := range []struct {
		name        string
		filename    string
		contentType string
		data        []byte
		want        string
		wantErr     string
	}{
		{name: "text extension is case insensitive", filename: "RESUME.TXT", data: []byte("Ada"), want: "Ada"},
		{name: "markdown media type", filename: "resume", contentType: "text/markdown", data: []byte("# Ada"), want: "# Ada"},
		{name: "PDF extension extracts text and links", filename: "resume.pdf", data: pdf, want: "Ada Lovelace\n\nhttps://example.test/portfolio"},
		{name: "PDF media type works without extension", filename: "resume", contentType: "application/pdf", data: pdf, want: "Ada Lovelace\n\nhttps://example.test/portfolio"},
		{name: "empty text is rejected", filename: "resume.txt", data: []byte(" \n\t "), wantErr: "no text content found in file"},
		{name: "unsupported extension", filename: "resume.docx", data: []byte("Ada"), wantErr: "unsupported file type: .docx"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractResumeText(tt.filename, tt.contentType, tt.data)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractResumeText() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("extractResumeText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractPDFTextRejectsInvalidOrTextlessPDF(t *testing.T) {
	for _, data := range [][]byte{
		[]byte("not a PDF"),
		testPDF("", ""),
	} {
		if _, err := extractPDFText(data); err == nil {
			t.Errorf("extractPDFText(%q) succeeded, want error", data)
		}
	}
}

func testPDF(text, url string) []byte {
	stream := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", text)
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R /Annots [6 0 R] >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		fmt.Sprintf("<< /Type /Annot /Subtype /Link /URI(%s) >>", url),
	}

	var pdf strings.Builder
	pdf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return []byte(pdf.String())
}
