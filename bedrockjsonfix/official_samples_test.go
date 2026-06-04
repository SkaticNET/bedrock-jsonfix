package bedrockjsonfix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOfficialBedrockSamplesResourcePackJSON(t *testing.T) {
	dir := os.Getenv("BEDROCK_SAMPLES_RESOURCE_PACK_DIR")
	if dir == "" {
		t.Skip("set BEDROCK_SAMPLES_RESOURCE_PACK_DIR to Mojang/bedrock-samples/resource_pack to run")
	}

	opt := DefaultOptions()
	opt.Pretty = false
	opt.PreserveIfValid = false

	checked := 0
	var failures []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, walkErr))
			return nil
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".json") {
			return nil
		}
		checked++
		input, readErr := os.ReadFile(path)
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, readErr))
			return nil
		}
		if _, fixErr := FixBytes(input, opt); fixErr != nil {
			rel, _ := filepath.Rel(dir, path)
			failures = append(failures, fmt.Sprintf("%s: %v", rel, fixErr))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatalf("no JSON files found under %s", dir)
	}
	if len(failures) > 0 {
		const maxFailures = 20
		if len(failures) > maxFailures {
			failures = failures[:maxFailures]
		}
		t.Fatalf("official Bedrock resource_pack JSON failures: %s", strings.Join(failures, "\n"))
	}
}
