package service

import (
	"testing"
	"time"
)

func TestBuildApplicationContainerName(t *testing.T) {
	createdAt := time.Date(2026, time.August, 28, 0, 1, 2, 0, time.FixedZone("CST", 8*60*60))
	tests := []struct {
		name      string
		applicant string
		want      string
	}{
		{name: "Chinese name", applicant: "张三", want: "zhangsan-2026-08-28-00-01-02"},
		{name: "compound surname", applicant: "欧阳娜娜", want: "ouyangnana-2026-08-28-00-01-02"},
		{name: "Latin name", applicant: "Zhang San", want: "zhangsan-2026-08-28-00-01-02"},
		{name: "mixed separators", applicant: "张-San_3", want: "zhangsan3-2026-08-28-00-01-02"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildApplicationContainerName(test.applicant, createdAt)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("buildApplicationContainerName(%q) = %q, want %q", test.applicant, got, test.want)
			}
		})
	}
}

func TestBuildApplicationContainerNameRejectsUnsupportedName(t *testing.T) {
	if _, err := buildApplicationContainerName("😀", time.Now()); err == nil {
		t.Fatal("expected unsupported applicant name to be rejected")
	}
}
