package internal

import (
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {

	dl := Download{
		Url:           "/testing",
		TotalSections: 10,
	}

	t.Run("should set targetPath correctly", func(t *testing.T) {
		expected := "targetPath"

		dl.setSavePath(expected)

		pathArr := strings.Split(dl.TargetPath, "/")

		got := pathArr[len(pathArr)-1]

		if expected != got {
			t.Errorf("expected %s got %v", expected, dl.TargetPath)
		}

	})

	t.Run("should set temp directory path correctly", func(t *testing.T) {
		expected := "/targetPath"

		dl.setTempDir(expected)

		if expected != dl.tmpDir {
			t.Errorf("expected %s got %v", expected, dl.tmpDir)
		}

	})

	t.Run("tmpfile array length should equal TotalSections", func(t *testing.T) {

		dl.setTmpFilesArray()

		got := len(dl.tmpFiles)

		if dl.TotalSections != got {
			t.Errorf("expected %d got %v", dl.TotalSections, got)
		}

	})

	t.Run("should compute byte sections correctly", func(t *testing.T) {

		size := 100000

		sections := dl.computeSections(size)

		sectionTests := []struct {
			name     string
			expected int
			got      int
		}{
			{name: "First Section", expected: 10000, got: sections[0][1]},
			{name: "Last Section", expected: 99999, got: sections[9][1]},
		}

		for _, tt := range sectionTests {
			t.Run(tt.name, func(t *testing.T) {

				if tt.expected != tt.got {
					t.Errorf("expected %d got %v", tt.expected, tt.got)
				}
			})
		}

	})

}
